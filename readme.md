### 数据库比对工具说明

##### 文件目录结构

```script
/
|-- check           主程序
|-- db.config       需要检查的数据配置
|-- xxx.sql         最新数据库结构，用Navicat导出的结构SQL
|-- xxx.modify.sql  脚本执行后输出的修改语句
```

#### 开始准备

##### 1.配置db.config

```script
用户名:密码@(服务器ip:端口)

eg: xxxx:xxxxx@(127.0.0.1:3306)
```

#### 使用介绍

```script
[linux]
./check [数据库结构文件]
eg：
./check ucenter

[windows]
打开命令行

./check.exe [数据库结构文件]
eg：
./check.exe ucenter

```
