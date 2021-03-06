FROM golang:1.9.4

# Add testing deps for curl
RUN echo 'deb http://httpredir.debian.org/debian testing main non-free contrib' >> /etc/apt/sources.list

# Grab deps (jq, hexdump, xxd, killall)
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  jq bsdmainutils vim-common psmisc netcat curl

# Setup tendermint repo
ENV REPO $GOPATH/src/github.com/bcbchain/tendermint
ENV GOBIN $GOPATH/bin
WORKDIR $REPO

# Install the vendored dependencies before copying code
# docker caching prevents reinstall on code change!
ADD Gopkg.toml Gopkg.toml
ADD Gopkg.lock Gopkg.lock
ADD Makefile Makefile
RUN make get_tools
RUN make get_vendor_deps

# Install the apps
ADD scripts scripts
RUN bash scripts/install_abci_apps.sh

# Now copy in the code
# NOTE: this will overwrite whatever is in vendor/
COPY . $REPO


RUN go install ./cmd/tendermint

# expose the volume for debugging
VOLUME $REPO

EXPOSE 46656
EXPOSE 46657
