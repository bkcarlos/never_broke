# Build the static Linux binary first; see FEATURE_DOCUMENTATION.md.
FROM busybox:1
WORKDIR /app
RUN mkdir -p logs keys && chown -R 10001:10001 /app
COPY --chown=10001:10001 neverbroke /app/neverbroke
USER 10001:10001
EXPOSE 8080
CMD ["/app/neverbroke"]
