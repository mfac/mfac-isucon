use strict;
use warnings;
use utf8;
use feature qw(say);

use String::Random;
use File::Slurp qw(read_file);
use List::Util qw(shuffle);
use Encode qw(encode_utf8);

my $sr = String::Random->new();
my @tags = split /\n/, read_file('tags.txt');


for (1..100000) {
    my $a = int rand(130) + 1;
    my $b = int rand(5) + 1;

    my $memo = $sr->randregex("[a-zA-Zあいうえおかきくけこさしすせそたちつてとなにぬねの]{$a}\\d{$b}");

    my $max = int rand(3);
    if ($max) {
        my @stags = shuffle @tags;
        my @t = @stags[0 .. $max];
        
        my $t = join " ", map { "#$_" } @t;
        $memo .= " $t";
    }
    say encode_utf8 $memo;
}
