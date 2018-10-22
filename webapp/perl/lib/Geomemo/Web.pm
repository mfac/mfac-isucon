package Geomemo::Web;
use 5.014;
use warnings;
use utf8;
use Kossy;
use DBIx::Sunny;
use URI::Escape qw/uri_escape_utf8/;
use Text::Xslate::Util qw/html_escape/;
use JSON::Types;

sub config {
    state $conf = {
        dsn           => $ENV{GEOMEMO_DSN}         // 'dbi:mysql:db=geomemo',
        db_user       => $ENV{GEOMEMO_DB_USER}     // 'root',
        db_password   => $ENV{GEOMEMO_DB_PASSWORD} // '',
    };
    my $key = shift;
    my $v = $conf->{$key};
    unless (defined $v) {
        die "config value of $key undefined";
    }
    return $v;
}

sub dbh {
    my ($self) = @_;
    return $self->{dbh} //= DBIx::Sunny->connect(config('dsn'), config('db_user'), config('db_password'), {
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

get '/initialize' => sub {
    my ($self, $c)  = @_;

    $self->dbh->query(q[
        DELETE FROM memo WHERE id > 100001
    ]);

    $c->render_json({
        result => 'ok',
    });
};

get '/' => sub {
    my ($self, $c)  = @_;

    $c->render('index.html');
};

get '/memo' => sub {
    my ($self, $c)  = @_;

    my $PER_PAGE = 20;
    my $page = $c->req->parameters->{page} || 1;

    my $limit = $PER_PAGE + 1;
    my $offset = $PER_PAGE * ($page - 1);

    my $memos = $self->dbh->select_all(q[
        SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        ORDER BY created_at DESC
        LIMIT :limit
        OFFSET :offset
    ], {
        limit => $limit,
        offset => $offset,
    });

    my $has_prev = $page > 1 ? 1 : 0;
    my $has_next = @$memos == $limit ? 1 : 0;

    # 余分にとった分を削除
    if ($has_next) {
        pop @$memos;
    }

    $memos = $self->inflate_memos($c, $memos);

    $c->render_json({
        memos     => $memos,
        page      => number $page,
        has_prev  => bool $has_prev,
        has_next  => bool $has_next,
    });
};

get '/tag/:tag' => sub {
    my ($self, $c) = @_;

    my $tag = $c->args->{tag};

    my $PER_PAGE = 20;
    my $page = $c->req->parameters->{page} || 1;

    my $limit = $PER_PAGE + 1;
    my $offset = $PER_PAGE * ($page - 1);

    my $memos = $self->dbh->select_all(q[
        SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        WHERE
          body REGEXP :regexp
        ORDER BY created_at DESC
        LIMIT :limit
        OFFSET :offset
    ], {
        regexp => sprintf('#%s[[:>:]]', $tag),
        limit  => $limit,
        offset => $offset,
    });

    my $has_prev = $page > 1 ? 1 : 0;
    my $has_next = @$memos == $limit ? 1 : 0;

    # 余分にとった分を削除
    if ($has_next) {
        pop @$memos;
    }

    $memos = $self->inflate_memos($c, $memos);

    $c->render_json({
        memos     => $memos,
        page      => number $page,
        has_prev  => bool $has_prev,
        has_next  => bool $has_next,
    });
};

get '/around/:memo_id' => sub {
    my ($self, $c) = @_;

    my $memo_id = $c->args->{memo_id};

    my $limit = 20;

    my $memo = $self->dbh->select_row(q[
        SELECT id, X(latlng) lat, Y(latlng) lng FROM memo
        WHERE
            id = :memo_id
    ], {
        memo_id => $memo_id,
    });

    my $sql = sprintf(q[
        SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        ORDER BY 
          GLength(
            GeomFromText(
                CONCAT('LineString(%f %f,', X(latlng),' ',Y(latlng), ')')
            )
          )
        LIMIT :limit
    ], $memo->{lat}, $memo->{lng});

    my $memos = $self->dbh->select_all($sql, { limit => $limit });
    $memos = $self->inflate_memos($c, $memos);

    $c->render_json({
        memos     => $memos,
        page      => number 1,
        has_prev  => bool 0,
        has_next  => bool 0,
    });
};

post '/memo' => sub {
    my ($self, $c) = @_;

    my $lat  = $c->req->parameters->{lat};
    my $lng  = $c->req->parameters->{lng};
    my $body = $c->req->parameters->{body};

    my $sql = sprintf(q[
        INSERT INTO memo
        (latlng, body, created_at, updated_at) VALUES 
        (GeomFromText('POINT(%f %f)'), :body, NOW(), NOW())
    ], $lat, $lng);

    $self->dbh->query($sql, { body => $body });
    my $memo_id = $self->dbh->last_insert_id;
    my $memo = $self->dbh->select_row(q[
        SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        WHERE id = ?
    ], $memo_id);

    my $memos = $self->inflate_memos($c, [$memo]);
    $memo = $memos->[0];

    $c->render_json($memo)
};

post '/memo/:memo_id/emoji/:emoji' => sub {
    my ($self, $c) = @_;

    $self->dbh->query(q[
        INSERT INTO memo_emoji 
        (memo_id, emoji, created_at) VALUES
        (:memo_id, :emoji, NOW())
    ], {
        memo_id => $c->args->{memo_id},
        emoji   => $c->args->{emoji},
    });

    $c->render_json({
        memo_id => number $c->args->{memo_id},
        emoji   => string $c->args->{emoji},
    });
};

sub inflate_memos {
    my ($self, $c, $memos) = @_;

    for my $memo (@$memos) {
        $memo->{emojis} = $self->dbh->select_all(q[
            SELECT emoji, COUNT(emoji) count
            FROM memo_emoji
            WHERE memo_id = :memo_id
            GROUP BY emoji
        ], {
            memo_id => $memo->{id}
        });
    }

    return $memos;
}

1;
