FROM golang:1.8

# Copy binary into image
COPY ./campaigns.json /
COPY ./adserver /

# Expose port
EXPOSE 8083

CMD ["--help"]
ENTRYPOINT ["/adserver"]

