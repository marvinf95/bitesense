# Frontend container is **nginx-only**: it expects the Flutter Web build to
# exist at `./web/` (produced on the dev notebook via `flutter build web` and
# rsynced to the Pi by the root Makefile). This mirrors the Goalbooze pattern
# from the Pi-Homelab and avoids carrying the multi-GB Flutter SDK image when
# building on the Pi via BuildKit.
#
# For external clusters that want a one-shot image build with Flutter inside,
# see Dockerfile.fullbuild.
FROM nginxinc/nginx-unprivileged:1.27-alpine
USER 0
RUN rm -rf /usr/share/nginx/html/*
COPY web/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
USER 101
EXPOSE 8080
