package Portal::Web;
use 5.028;
use warnings;
use utf8;
use Kossy;
use Portal::Container;

sub dbh {
    my $self = shift;
    $self->{dbh} //= container('db')->dbh
}

filter 'auth' => sub {
    my $app = shift;
    sub {
        my ($self, $c) = @_;
        my $team_id = $c->env->{'psgix.session'}->{team_id};
        $c->halt(403) unless defined $team_id;

        $c->stash->{team_id} = $team_id;
        $app->($self, $c);
    }
};

filter 'check_team' => sub {
    my $app = shift;
    sub {
        my ($self, $c) = @_;
        my $team_id = $c->env->{'psgix.session'}->{team_id};
        return $c->redirect('/login') unless defined $team_id;

        $app->($self, $c);
    }
};


get '/login' => sub {
    my ($self, $c) = @_;
    $c->render('login.tx');
};

post '/login' => sub {
    my ($self, $c) = @_;

    my $name = $c->req->parameters->{name};
    my $pass = $c->req->parameters->{password};
    my $team_id = $self->dbh->select_one(q[
        SELECT id FROM teams
        WHERE name = ? AND password = ?
    ], $name, $pass);

    $c->env->{'psgix.session'}->{team_id} = $team_id;
    $c->redirect('/');
};

post '/logout' => sub {
    my ($self, $c) = @_;
    $c->env->{'psgix.session'} = {};
    $c->redirect('/');
};

get '/' => [qw/check_team/] => sub {
    my ($self, $c)  = @_;
    $c->render('index.html');
};

get '/team' => [qw/auth/] => sub {
    my ($self, $c) = @_;

    my $team_id = $c->stash->{team_id};
    my $team = $self->dbh->select_row(q[
        SELECT id, name FROM teams
        WHERE id = ?
    ], $team_id);

    $c->render_json($team);
};

get '/pass_scores' => sub {
    my ($self, $c)  = @_;

    my $pass_scores = $self->dbh->select_all(q[
        SELECT score, team_id, created_at FROM scores
        WHERE pass = TRUE
        ORDER BY created_at DESC
    ]);

    my $teams = $self->dbh->selectall_hashref(q[
        SELECT id, name, color FROM teams
    ], 'id');

    # TODO: client で組み立てる
    my %data;
    for my $ps (@$pass_scores) {
        my $team_id = $ps->{team_id};
        my $team = $teams->{$team_id};
        $data{$team_id} //= {
            label           => $team->{name},
            backgroundColor => $team->{color},
            borderColor     => $team->{color},
            data            => [],
        };

        push @{$data{$team_id}->{data}} => {
            t => $ps->{created_at},
            y => $ps->{score},
        };
    }

    $c->render_json([values %data]);
};

get '/ranking_scores' => sub {
    my ($self, $c) = @_;

    my $rankings = $self->dbh->select_all(q[
        SELECT team_id, MAX(score) max_score FROM scores
        GROUP BY team_id
        ORDER BY max_score DESC
    ]);

    my $teams = $self->dbh->selectall_hashref(q[
        SELECT id, name, color FROM teams
    ], 'id');

    # TODO: client で組み立てる
    my @data;
    for my $r (@$rankings) {
        my $team = $teams->{$r->{team_id}};
        push @data => {
            team_name => $team->{name},
            max_score => $r->{max_score},
        }
    }

    $c->render_json(\@data);
};

get '/team_scores' => [qw/auth/] => sub {
    my ($self, $c) = @_;

    my $team_id = $c->stash->{team_id};
    my $scores = $self->dbh->select_all(q[
        SELECT pass, score, message, created_at FROM scores
        WHERE team_id = ?
        ORDER BY created_at DESC
    ], $team_id);

    $c->render_json($scores);
};

get '/jobs' => [qw/auth/] => sub {
    my ($self, $c) = @_;

    my $jobs = $self->dbh->select_all(q[
        SELECT team_id, status, enqueued_at FROM jobs
        WHERE status != 'done'
    ]);

    my $teams = $self->dbh->selectall_hashref(q[
        SELECT id, name FROM teams
    ], 'id');

    my @data;
    for my $job (@$jobs) {
        my $team = $teams->{$job->{team_id}};
        push @data => {
            team_name => $team->{name},
            enqueued_at => $job->{enqueued_at},
            status => $job->{status},
        };
    }

    $c->render_json(\@data);
};

post '/enqueue' => [qw/auth/] => sub {
    my ($self, $c) = @_;
    my $team_id = $c->stash->{team_id};
    my $team = $self->dbh->select_row(q[
        SELECT name FROM teams
        WHERE id = ?
    ], $team_id);

    my $already_enqueued = $self->dbh->select_one(q[
        SELECT id FROM jobs
        WHERE status != 'done'
        AND team_id = ?
    ], $team_id);

    if ($already_enqueued) {
        return $c->render_json({
            result => 'fail',
            reason => 'Benchmark already enqueued',
        });
    }

    $self->dbh->query(q[
        INSERT INTO jobs
        (team_id, status, enqueued_at, updated_at, created_at) VALUES
        (?, 'waiting', NOW(), NOW(), NOW())
    ], $team_id);
    my $job_id = $self->dbh->last_insert_id;

    container('resque')->push(config('bench_queue') => {
        class => 'Portal::Bench',
        args => [$team_id, $job_id],
    });

    $c->render_json({
        result => 'success',
        reason => '',
    });
};

1;
