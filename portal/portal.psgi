#!/usr/bin/env plackup
use 5.014;
use warnings;

use FindBin;
use lib "$FindBin::Bin/lib";
use File::Spec;
use Plack::Builder;

use Portal::Web;

my $root_dir = $FindBin::Bin;
my $app = Portal::Web->psgi($FindBin::Bin);
builder {
    enable 'ReverseProxy';
    enable 'Static',
        path => qr!^/static/(?:(?:css|js|img)/|favicon\.ico$)!,
        root => File::Spec->catfile($root_dir, qw(views));
    enable 'Session::Cookie',
        session_key => "mfac_isucon_portal_session",
        secret      => 'tonymoris';
    $app;
};
