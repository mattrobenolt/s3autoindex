FROM scratch
ADD bin/s3autoindex s3autoindex
EXPOSE 8000
ENTRYPOINT ["/s3autoindex"]
