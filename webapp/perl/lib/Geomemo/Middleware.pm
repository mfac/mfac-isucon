package Geomemo::Middleware;
use 5.014;
use warnings;

use parent 'Plack::Middleware';

sub call {
    my ($self, $env) = @_;

    # 全部JSON API
    $env->{'kossy.request.parse_json_body'} = !!1;
    
    return $self->app->($env)
}


1;
