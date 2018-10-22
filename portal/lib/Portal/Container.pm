package Portal::Container;
use 5.028;
use warnings;
use utf8;

use DBIx::Sunny;
use DBIx::Handler;
use Resque;

use parent qw(Exporter);

our @EXPORT = qw(container config);

sub container {
    my $name = shift;
    my $code = __PACKAGE__->can($name);
    $code ? $code->(@_) : die "cannot call $name";
}

sub meta {
    my $class = shift;
    state $meta = {};
    return $meta;
}

sub config {
    state $conf = {
        dsn            => $ENV{MFAC_ISUCON_PORTAL_DSN}         // 'dbi:mysql:db=mfac_isucon_portal',
        db_user        => $ENV{MFAC_ISUCON_PORTAL_DB_USER}     // 'root',
        db_password    => $ENV{MFAC_ISUCON_PORTAL_DB_PASSWORD} // '',

        redis_server   => $ENV{MFAC_ISUCON_PORTAL_REDIS_SERVER} // '127.0.0.1:6379',

        bench          => $ENV{MFAC_ISUCON_PORTAL_BENCH} // "/home/isucon/mfac-isucon/bench/bin/bench",
        bench_userdata => $ENV{MFAC_ISUCON_PORTAL_BENCH_USERDATA} // "/home/isucon/mfac-isucon/bench/userdata/",
        bench_queue    => 'queue',
    };

    my $key = shift;
    my $v = $conf->{$key};
    unless (defined $v) {
        die "config value of $key undefined";
    }
    return $v;
}


sub db {
    __PACKAGE__->meta->{db} //= DBIx::Handler->new(config('dsn'), config('db_user'), config('db_password'), {
        RootClass => 'DBIx::Sunny',
        Callbacks => {
            connected => sub {
                my $dbh = shift;
                $dbh->do(q[SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY']);
                $dbh->do('SET NAMES utf8mb4');
                return;
            },
        },
    });
}

sub resque {
    __PACKAGE__->meta->{resque} //= Resque->new(redis => config('redis_server'));
}


1;
