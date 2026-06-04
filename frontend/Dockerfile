# Frontend container is **nginx-only**: it expects `flutter build web` to have
# run already, producing `build/web/`. The root Makefile orchestrates the build
# on the dev notebook and rsyncs the result to the Pi for `pibuild`.
#
# For external clusters that want a one-shot build with Flutter inside the
# image, see Dockerfile.fullbuild.
FROM nginxinc/nginx-unprivileged:1.27-alpine
USER 0
RUN rm -rf /usr/share/nginx/html/*
COPY build/web/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
USER 101
EXPOSE 8080
