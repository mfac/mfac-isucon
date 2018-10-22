use strict;
use warnings;
use utf8;
use feature qw(say);

use DBIx::Sunny;
use File::Slurp qw(read_file);
use List::Util qw(shuffle);

my $dbh = DBIx::Sunny->connect('dbi:mysql:db=geomemo', 'root', '');
my @memos = split /\n/, read_file('memos.txt');
my @emojis = split /\n/, read_file('emojis.txt');

my $sql = q[
INSERT INTO memo (id, latlng, body, created_at, updated_at) VALUES 
];

my $sql_emoji = q[
INSERT INTO memo_emoji (memo_id, emoji, created_at) VALUES
];

my $id = 1;
my @sql;
for my $body (@memos) {
    my $lat = rand(90);
    my $lng = rand(180);

    push @sql => "($id, GeomFromText('POINT($lat $lng)'), '$body', NOW(), NOW())";
    if ($id % 1000 == 0) {
        say $sql . join(",\n", @sql) . ";\n\n";
        @sql = ();
    }

    my $max = int rand(3);
    if ($max) {
        my @semojis = shuffle @emojis;
        my @e = @semojis[0 .. $max];
        my $count = int rand($max * 3) + 1;

        my @sql_emoji;
        for my $emoji (@e) {
            for (1 .. $count) {
                push @sql_emoji => sprintf("(%d, '%s', NOW())", $id, $emoji)
            }
        }
        say $sql_emoji . join(",\n", @sql_emoji) . ";\n\n"
    }
    $id++;
}
