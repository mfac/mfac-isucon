#!/usr/bin/env perl
use 5.014;
use warnings;

use FindBin;
use lib "$FindBin::Bin/lib";
use Portal::Container;
use Log::Minimal;

my $worker = container('resque')->worker;
$worker->add_queue(config('bench_queue'));

infof('start worker');
$worker->work;
infof('end worker');
