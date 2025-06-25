# opensearch  运行方式 失败
    docker run -d \
        --name opensearch \
        --env cluster.name=opensearch-cluster \
        --env node.name=opensearch-node1 \
        --env discovery.seed_hosts=127.0.0.1 \
        --env "OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m" \
        -p 9200:9200 -p 9600:9600 \
        -v /tmp/opensesarch/data:/usr/share/opensearch/data \
        opensearchproject/opensearch