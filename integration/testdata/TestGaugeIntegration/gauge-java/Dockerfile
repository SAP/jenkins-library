# Building the image
# 	docker build -t gauge-java .
# Running the image
# 	docker run  --rm -it -v ${PWD}/reports:/gauge/reports gauge-java

# This image uses the official openjdk base image.

FROM openjdk

# Install gauge
RUN microdnf install -y unzip \
    && curl -Ssl https://downloads.gauge.org/stable | sh

# Set working directory
WORKDIR /gauge
 
# Copy the local source folder
COPY . .

# Install gauge plugins
RUN gauge install \
    && gauge install screenshot

CMD ["gauge", "run",  "specs"]
