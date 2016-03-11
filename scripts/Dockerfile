FROM cassandra:latest
RUN sed -i 's/start_rpc: false/start_rpc: true/g' /etc/cassandra/cassandra.yaml
RUN cat /etc/cassandra/cassandra.yaml | grep "start_rpc:"
RUN sed -i 's/#MX4J_PORT="-Dmx4jport=8081"/MX4J_PORT="-Dmx4jport=8082"/g' /etc/cassandra/cassandra-env.sh
RUN sed -i 's/#MX4J_ADDRESS="-Dmx4jaddress=127.0.0.1"/MX4J_ADDRESS="-Dmx4jaddress=0.0.0.0"/g' /etc/cassandra/cassandra-env.sh
RUN sed -i 's/LOCAL_JMX=yes/LOCAL_JMX=no/g' /etc/cassandra/cassandra-env.sh
RUN sed -i 's/-Dcom.sun.management.jmxremote.authenticate=true/-Dcom.sun.management.jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh
RUN cat /etc/cassandra/cassandra.yaml | grep "start_rpc:"

RUN  apt-get update \
  && apt-get install -y wget \
  && rm -rf /var/lib/apt/lists/*

RUN cd /tmp && wget "http://downloads.sourceforge.net/project/mx4j/MX4J%20Binary/3.0.2/mx4j-3.0.2.tar.gz" \
 && tar zxf mx4j-3.0.2.tar.gz && cp  mx4j-3.0.2/lib/mx4j-tools.jar /usr/share/cassandra/lib/

EXPOSE 8082