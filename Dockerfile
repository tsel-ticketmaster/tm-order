# Image Builder
FROM telkomindonesia/alpine:go-1.21 AS go-builder

LABEL maintainer="patrick_m_sangian@telkomsel.co.id"

# Set Working Directory
WORKDIR /usr/src/app

# Copy Source Code
COPY . ./

# Dependencies installation and binary file builder
RUN apk add --no-cache make

RUN make install \
  && make build


# Final Image
# ---------------------------------------------------
FROM dimaskiddo/alpine:base

# Set Working Directory
WORKDIR /usr/src/app

# Copy Anything The Application Needs
COPY --from=go-builder /tmp/app ./

# Expose Application Port
EXPOSE 9000

# Run The Application
CMD ["./app"]