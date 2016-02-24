[![Build Status](https://travis-ci.org/ngaut/codis-ha.svg?branch=master)](https://travis-ci.org/ngaut/codis-ha)


Usage:

go get github.com/ngaut/codis-ha


cd codis-ha


go build

# 编辑json格式的配置文件，输入自己的配置，参考示例codis-ha.json
# dashboard_addr : codis-config dashboard 地址
# product_name : 产品名称,同codis-config中的 product 一致
# log_file : 日志文件名称
# log_level : 日志等级 debug/info/error
# check_interval : 每次检测的时间间隔
# max_try_times : PING失败多少次认为server故障
# email_addr : 发送邮件使用的账号
# email_pwd : 发送邮件使用的密码
# smtp_addr : 邮件SMTP服务器地址端口
# to_addr : 告警邮件发送接收地址列表，多个接收人账号之间以分号分隔
# send_interval : 发送邮件间隔时间，单位是秒(避免server故障后未及时恢复导致频繁发送邮件)

codis-ha [codis-ha.json]


