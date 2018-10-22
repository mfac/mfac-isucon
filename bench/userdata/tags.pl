use strict;
use warnings;
use utf8;
use feature qw(say);

use String::Random;

my $sr = String::Random->new();


for (1..1000) {
    my $a = int rand(5) + 1;
    my $b = int rand(5) + 1;

    say $sr->randregex("[a-zA-Z]{$a}\\d{$b}");
}
