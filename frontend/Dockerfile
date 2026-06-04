# Build stage: Flutter web release.
FROM ghcr.io/cirruslabs/flutter:3.24.3 AS build
WORKDIR /app
COPY pubspec.yaml pubspec.lock* ./
RUN flutter pub get
COPY . .
RUN flutter build web --release --base-href / \
    --dart-define=BITESENSE_API=__API_PLACEHOLDER__

# Serve stage: nginx unprivileged.
FROM nginxinc/nginx-unprivileged:1.27-alpine
USER 0
RUN rm -rf /usr/share/nginx/html/*
COPY --from=build /app/build/web /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
USER 101
EXPOSE 8080
