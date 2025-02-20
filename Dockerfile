FROM debian:bookworm
RUN apt-get update && \
  apt-get install -y ffmpeg cmake libfftw3-dev git

RUN apt-get update && apt-get install -y \
  wget \
  libmad0 \
  libid3tag0 \
  libgd3 \
  libboost-regex-dev \
  libboost-program-options-dev \
  libboost-filesystem-dev

ARG TARGETARCH
RUN wget "https://github.com/bbc/audiowaveform/releases/download/1.10.1/audiowaveform_1.10.1-1-12_${TARGETARCH}.deb" && \
  dpkg -i "audiowaveform_1.10.1-1-12_${TARGETARCH}.deb" && \
  rm "audiowaveform_1.10.1-1-12_${TARGETARCH}.deb"
RUN mkdir -p /tmp/
RUN mkdir -p /tmp/build
RUN mkdir -p /tmp/libkeyfinder
RUN git clone https://github.com/mixxxdj/libkeyfinder.git /tmp/libkeyfinder
WORKDIR /tmp/libkeyfinder
RUN apt-get install -y gcc-12 g++-12
ENV CC=gcc-12
ENV CXX=g++-12
RUN mkdir -p /tmp/libkeyfinder/build
RUN cmake -S . -B /tmp/libkeyfinder/build -DBUILD_TESTING=OFF
RUN cmake --build /tmp/libkeyfinder/build 
RUN cmake --install /tmp/libkeyfinder/build
RUN mkdir -p /tmp/keyfinder-cli
RUN git clone https://github.com/evanpurkhiser/keyfinder-cli.git /tmp/keyfinder-cli
WORKDIR /tmp/keyfinder-cli
RUN apt-get install -y libavutil-dev libavcodec-dev libavformat-dev libswresample-dev
ENV LD_LIBRARY_PATH=/usr/local/lib/:/usr/lib/
RUN make && make install
RUN apt-get install -y python3-pkg-resources python3-numpy
RUN apt-get install -y aubio-tools python3-aubio
RUN wget  https://exiftool.org/Image-ExifTool-13.19.tar.gz && \
  gzip -dc Image-ExifTool-13.19.tar.gz | tar -xf - && \
  cd Image-ExifTool-13.19 && \
  perl Makefile.PL && \
  make install
RUN rm -rf /tmp/
RUN mkdir -p /app/
WORKDIR /app/