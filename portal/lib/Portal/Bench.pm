package Portal::Bench;
use 5.028;
use warnings;
use utf8;

use Portal::Container;
use Try::Tiny;
use JSON::XS qw(decode_json);
use Capture::Tiny qw(capture);
use Log::Minimal;

my $TIMEOUT = 100;

sub perform {
    my $job = shift;

    infof 'start Portal::Bench';

    my $team_id = $job->args->[0];
    my $job_id = $job->args->[1];

    container('db')->dbh->query(q[
        UPDATE jobs SET status = 'running'
        WHERE id = ?
    ], $job_id);

    my $app_host = container('db')->dbh->select_one(q[
        SELECT app_host FROM teams
        WHERE id = ?
        LIMIT 1
    ], $team_id);

    my $result;
    try {
        local $SIG{ALRM} = sub { die 'TIMEOUT' };
        alarm($TIMEOUT);
        $result = execute_bench($app_host);
    }
    catch {
        my $e = shift;
        if ($e eq 'TIMEOUT') {
            $result = {
                pass    => 0,
                score   => 0,
                message => 'TIMEOUT'
            }
        }
        else {
            critf "unexpected error: $e";
        }
    };

    try {
        container('db')->dbh->query(q[
            UPDATE jobs SET status = 'done'
            WHERE id = ?
        ], $job_id);

        my $message = $result->{message};
        my $submessage = substr($message, 0, 300);

        container('db')->dbh->query(q[
            INSERT INTO scores
            (pass, score, message, team_id, updated_at, created_at) VALUES
            (?, ?, ?, ?, NOW(), NOW())
        ], @$result{qw(pass score)}, $submessage, $team_id);
    }
    catch {
        my $e = shift;
        critf $e;
    };

    infof 'end Portal::Bench';
}

sub execute_bench {
    my $app_host = shift;

    my ($stdout, $stderr, $exit) = capture {
        system( config('bench'), '-userdata', config('bench_userdata'), '-target', $app_host);
    };

    if ($exit != 0) {
        return {
            pass    => 0,
            score   => 0,
            message => sprintf("Bench failed. %s", $stdout),
        }
    }

    # e.g. {"pass":false,"score":0,"success":0,"fail":0,"messages":["初期化リクエストに失敗しました"]}
    
    my $data;
    try {
        $data = decode_json($stdout);
    }
    catch {
        my $e = shift;
        return {
            pass    => 0,
            score   => 0,
            message => sprintf("Bench failed. decode_json error: %s, benchmark stdout: %s", $e, $stdout),
        }
    };

    if (!$data || !%$data) {

        warnf $stdout;
        warnf $stderr;

        return +{
            pass => 0,
            score => 0,
            message => sprintf("Bench failed. benchmarker stdout: %s, stderr: %s", $stdout, $stderr),
        }
    }

    return +{
        pass    => $data->{pass},
        score   => $data->{score},
        message => join ' ', @{$data->{messages}||[]},
    }
}

1;
