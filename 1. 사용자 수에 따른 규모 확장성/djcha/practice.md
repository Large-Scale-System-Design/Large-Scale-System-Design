# 2주차. MariaDB 다중화 VIP 설정 (실습)

<br>

## OS: Rocky9.6

<br>

# 1\. Master

Network Setting

```
sudo nmcli connection modify ens160 ipv4.method manual ipv4.addresses 172.16.43.101/24 ipv4.gateway 172.16.43.2 ipv4.dns "8.8.8.8,1.1.1.1"
sudo nmcli connection up ens160
```

<br>

MariaDB 설치

```
sudo dnf install mariadb-server -y
sudo systemctl enable --now mariadb  # 서비스 활성화 및 즉시 시작
```

<br>

방화벽 설정

```
sudo firewall-cmd --permanent --add-service=mysql
sudo firewall-cmd --reload
```

<br>

MariaDB 주소 허용

```
sudo vi /etc/my.cnf.d/mariadb-server.cnf

bind-address -> 0.0.0.0 으로 허용
```

<br>

MariaDB 재시작

```
sudo systemctl restart mariadb
```

<br>

MariaDB 설정

```
vi /etc/my.cnf.d/mariadb-server.cnf

sudo mkdir -p /var/log/mysql
sudo chown mysql:mysql /var/log/mysql
```

<br>

![](Files/image.png)<br>

<br>

복제용 계정

```
MariaDB > 
CREATE USER 'replicator'@'%' IDENTIFIED BY 'password';
GRANT REPLICATION SLAVE ON *.* TO 'replicator'@'%';
FLUSH PRIVILEGES;
```

<br>

마스터 계정 상태 확인

![](Files/image%202.png)<br>

<br>

FIle, Position 값 잘 기억하기  

<br>

<br>

* * *

<br>

# 2\. Slave1

Network Setting

```
sudo nmcli connection modify ens160 ipv4.method manual ipv4.addresses 172.16.43.102/24 ipv4.gateway 172.16.43.2 ipv4.dns "8.8.8.8,1.1.1.1"
sudo nmcli connection up ens160
```

<br>

```
vi /etc/my.cnf.d/mariadb-server.cnf
```

<br>

![](Files/image%203.png)<br>

<br>

```
mysql -u root -p

MariaDB >
STOP SLAVE;
CHANGE MASTER TO
    MASTER_HOST='172.16.43.101',           -- Master 서버 IP
    MASTER_USER='replicator',              -- 복제용 사용자 이름
    MASTER_PASSWORD='my_password',       -- 복제용 사용자 비밀번호
    MASTER_LOG_FILE='mysql-bin.000001',    -- Master 상태에서 확인한 File 값
    MASTER_LOG_POS=328;                    -- Master 상태에서 확인한 Position 값
START SLAVE;
```

<br>

<br>

```
STOP SLAVE;
CHANGE MASTER TO
    MASTER_HOST='172.16.43.101',
    MASTER_USER='replicator',
    MASTER_PASSWORD='my_password',
    MASTER_LOG_FILE='mysql-bin.000001',
    MASTER_LOG_POS=669;
START SLAVE;
```

<br>

### 상태 체크 시 오해할만한 상황

아래 상태 확인하면 정상 상태인 것으로 확인

Slave\_IO\_State: waiting for master to send event → 이거는 잘 안된건 아님

- `Slave_IO_Running: Yes`
- `Slave_SQL_Running: Yes`

<br>

![](Files/image%204.png)<br>

<br>

<br>

<br>

* * *

<br>

# 3\. Slave2

Network Setting

```
sudo nmcli connection modify ens160 ipv4.method manual ipv4.addresses 172.16.43.103/24 ipv4.gateway 172.16.43.2 ipv4.dns "8.8.8.8,1.1.1.1"
sudo nmcli connection up ens160
```

<br>

```
vi /etc/my.cnf.d/mariadb-server.cnf
```

<br>

![](Files/image%205.png)<br>

<br>

```
mysql -u root -p

MariaDB >
STOP SLAVE;
CHANGE MASTER TO
    MASTER_HOST='192.168.1.101',           -- Master 서버 IP
    MASTER_USER='replicator',              -- 복제용 사용자 이름
    MASTER_PASSWORD='your_password',       -- 복제용 사용자 비밀번호
    MASTER_LOG_FILE='mysql-bin.000001',    -- Master 상태에서 확인한 File 값
    MASTER_LOG_POS=342;                    -- Master 상태에서 확인한 Position 값
START SLAVE;
```

<br>

<br>

<br>

<br>

# 4\. VIP 설정

```
sudo dnf install -y keepalived
```

<br>

<br>

```
sudo vi /etc/keepalived/keepalived.conf

global_defs {
   router_id DB_MASTER
}

vrrp_instance VI_1 {
    state MASTER             # 이 서버를 MASTER로 설정
    interface ens160         # 1단계에서 확인한 네트워크 인터페이스 이름
    virtual_router_id 51     # VRRP 그룹 ID, 모든 서버가 동일해야 함
    priority 150             # 우선순위, MASTER가 가장 높아야 함
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
        172.16.43.104/24     # 사용할 가상 IP (VIP)
    }
}

# Slave1
global_defs {
   router_id DB_SLAVE1
}

vrrp_instance VI_1 {
    state BACKUP              # 이 서버는 BACKUP
    interface ens160          # Master와 동일한 인터페이스 이름
    virtual_router_id 51      # Master와 동일한 ID
    priority 100              # Master보다 낮은 우선순위
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
        172.16.43.104/24      # Master와 동일한 VIP
    }
}

# Slave2
global_defs {
   router_id DB_SLAVE2
}

vrrp_instance VI_1 {
    state BACKUP              # 이 서버는 BACKUP
    interface ens160          # Master와 동일한 인터페이스 이름
    virtual_router_id 51      # Master와 동일한 ID
    priority 90              # Slave 1 보다 우선순위 낮게 해서 장애 순서 조치
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
        172.16.43.104/24      # Master와 동일한 VIP
    }
}
```

<br>

```
global_defs {
   router_id DB_MASTER
}

vrrp_instance VI_1 {
    state MASTER             
    interface ens160         
    virtual_router_id 51     
    priority 150             
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
        172.16.43.104/24
    }
}
```

<br>

![](Files/image%206.png)![](Files/image%207.png)<br>

<br>

Keepalived 활성화 + 상태 확인

```
sudo systemctl enable --now keepalived

ip a
```

<br>

Slave 서버에는 VIP 가 보이면 안됨.

만약 보인다라고 하면, 방화벽 SELINUX 꺼야함

<br>

```
systemctl disable --now firewalld
setenforce 0 # 임시
```

<br>

![](Files/image%208.png)<br>

![](Files/image%209.png)<br>

![](Files/image%2010.png)<br>

# 테스트

Master > dup\_db 데이터베이스에 샘플 테이블 생성

![](Files/image%2011.png)<br>

Master > 데이터 생성

Slave > user 테이블 확인

![](Files/image%2012.png)<br>

<br>

# Slave → Master 설정 다시 바꾸는 법

```
CHANGE MASTER TO MASTER_HOST='172.16.43.101', MASTER_USER='replicator', MASTER_PASSWORD='my_passsword', MASTER_LOG_FILE='mysql-bin.000001', MASTER_LOG_POS=669;

# SLAVE 재시작
START SLAVE;
```

### position 번호 다시 셋팅

![](Files/image%2013.png)<br>

<br>

<br>

# 삽질 모음

- DB, Table 데이터 싱크 맞춰놓고 해야됨.
    - Master엔 DB, Table 있고 Slave들에는 DB 없이 싱크맞췄다가, 백업안됨.
        - 그래서 다시 SLAVE 리셋하고 Position 번호 맞추고 DB 데이터 싱크 맞추고 삽질함

- VIP 설정 후 Master, Slave1, Slave2 모두 VIP가 잡혀있길래, 띠용 했는데, keepalived 관련한 통신을 서로 해야되는데
    - 방화벽에 의해 막혀서 서로 VIP 가 지꺼인줄 알고 욕심부림.
        - 방화벽 꺼서 서로 원만한 통신으로 합의보게 하니까 VIP 사라짐(우선순위 기준)

<br>

<br>