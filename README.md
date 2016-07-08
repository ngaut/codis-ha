[![Build Status](https://travis-ci.org/ngaut/codis-ha.svg?branch=master)](https://travis-ci.org/ngaut/codis-ha)

基于codis 2.x (https://github.com/wlibo666/codis)
Usage:

go get github.com/wlibo666/codis-ha

cd codis-ha

go build

### codis-ha.json 配置项说明
\# 编辑json格式的配置文件，输入自己的配置，参考示例codis-ha.json  
\# dashboard_addr : codis-config dashboard 地址  
\# product_name : 产品名称,同codis-config中的 product 一致  
\# log_file : 日志文件名称  
\# log_level : 日志等级 debug/info/error  
\# check_interval : 每次检测的时间间隔  
\# max_try_times : PING失败多少次认为server故障  
\# email_addr : 发送邮件使用的账号  
\# email_pwd : 发送邮件使用的密码  
\# smtp_addr : 邮件SMTP服务器地址端口  
\# to_addr : 告警邮件发送接收地址列表，多个接收人账号之间以分号分隔  
\# send_interval : 发送邮件间隔时间，单位是秒(避免server故障后未及时恢复导致频繁发送邮件)
\# master_save : master redis是否存储rdb文件，如果为空，则不存储，否则按照配置存储，如  600 1,则代表600秒只要有一个改变就存储备份文件
\# slave_save : slave redis是否存储rdb文件，如果为空，则不存储，否则按照配置存储，如  600 1,则代表600秒只要有一个改变就存储备份文件
    添加save配置的原因是 我们的业务为了效率要求master不存储rdb文件，slave必须存储rdb文件，这样主从切换后能保证master不进行rdb备份
    同时，我们有一个脚本每隔一个小时统一采集所有slave的rdb文件进行备份，防止salve磁盘损坏造成数据丢失。

codis-ha [codis-ha.json]


