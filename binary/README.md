# 用来搭建二进制集群如何搭建和参数调优问题
## 下载二进制包  一般安装到 /usr/local/src 下
    1：wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.17.24-linux-x86_64.tar.g
        tar -xzf elasticsearch-7.17.24-linux-x86_64.tar.gz
    2：集群配置一下免密.  ssh-key-gen ssh-copy-id 
        ssh-keygen -t rsa
        ssh-copy-id 192.168.1.1
    3：配置集群 ntp 同步
        apt-get install ntp || yum install ntp
        修改 /etc/ntp.conf, 开启集群时间同步
    4：创建数据存储目录, 日志存储目录，和启动账户。
        mkdir /data/esdata/
        mkdir /var/log/es 
        useradd elastic
        usermod -G root elastic
        chown -R elastic:root /data/esdata/ /var/log/es 

    5：配置设置密钥。注意，全部回车，不要特殊设置什么。 推荐放到 config 下面。 详细可见 yml 配置。
        ./elasticsearch-certutil ca
        ./elasticsearch-certutil cert --ca elastic-stack-ca.p12

    6：创建配置文件，并启动 elastic 配置
        创建启动服务的command
        /usr/lib/systemd/system/es.service

        需要改动的几处 memlock 位置
        https://stackoverflow.com/questions/45008355/elasticsearch-process-memory-locking-failed

        journal 来查日志

    7：kibana 的搭建，见notion

    other:
        生成密码 ./elasticsearch-setup-passwords auto 只能生成一次，且不能改密码。要注意保存

## es 的搭建步骤
    完成教程：https://www.cnblogs.com/lixinliang/p/17599228.html

## file handler 和 map concurrent 控制
    /etc/sysctl.conf
        fs.file-max = 2097152
        vm.max_map_count = 262144
    sudo sysctl -p