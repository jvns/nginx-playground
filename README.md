# nginx playground

This code is running at [nginx-playground.wizardzines.com](https://nginx-playground.wizardzines.com).
There's a [blog post](https://jvns.ca/blog/2021/09/24/new-tool--an-nginx-playground/) that
explains the design.

## mostly unmaintained

I haven't made any changes to this code for a couple of years and I don't
intend to add new features.

I'm open to merging bug fixes though, and if you find a bug I'd like to hear
about it in the issues. 

## things you might want to change

Some things that are specific to my deployment of this code, that you'll
want to remove or change if you make significant changes:

* "a wizard zines project" in the header
* the analytics (grep for `plausible`)
* the `fly.toml`
* the FAQ

## security notes

Might have security vulnerabilities, it gives the user a lot of access, I
personally would only run this software on a machine that I was comfortable
with potentially being compromised. I haven't had any issues that I know of yet
though.

I run the backend and the frontend on separate servers so that if the backend
ever did get compromised, it wouldn't affect the frontend site. You can just
run the frontend on GitHub pages or something.

## development setup

Developing the frontend is straightforward:

```
cd static
python3 -m http.server 8084 # or serve a webserver any other way you want
```

It has the URL of the backend hardcoded so you I work on it without having to run the backend.

To generate the Tailwind CSS: (with the [tailwind standalone CLI](https://tailwindcss.com/blog/standalone-cli))

```
tailwindcss-macos-arm64 --content  'static/*.html,static/*.js' -o static/css/tailwind-classes.css
```

Developing the backend is a bit of a mess. It only works on Linux (because it
depends on bubblewrap), and it's a pain to develop because it requires nginx to
be installed in specific directories and it has to run as root.

Here are setup instructions that worked for me on a fresh Ubuntu 22.10 install:

```bash
sudo apt-get install golang nginx bubblewrap
# install go-httpbin
go install github.com/mccutchen/go-httpbin/v2/cmd/go-httpbin@latest
# just putting go-httpbin in your PATH doesn't work
cp /the/path/to/go-httpbin /usr/bin 
# there has to be an nginx user
useradd -s /bin/false nginx 
# start the server
bash scripts/run-local.sh
```

You can test that it's working with `httpie` like this:

```
http post localhost:8080/run nginx_config=@examples/basic.conf command="http get localhost"
```

There's also a `watch-local.sh` that watches your changes with `entr`.

It should in theory be possible to modify the way `bubblewrap` is invoked to
run without being root (and have a less scary development experience), but I
couldn't figure out how to do it. If you figure out how to make the experience
of working with bubblewrap easier I'd love to hear about it.

Also you could definitely modify this so that it works with a separate nginx
container image and doesn't require you to have nginx installed in your main
filesystem.

