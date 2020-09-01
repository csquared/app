我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# convox/app

<a href="https://travis-ci.org/convox/app">
  <img align="right" src="https://travis-ci.org/convox/app.svg?branch=master">
</a>

Create a CloudFormation template from an app manifest.

This is a guide to developing the convox/app project. For detailed
installation and usage instructions, see [http://docs.convox.com/](http://docs.convox.com/).

## Development

```bash
$ go get github.com/convox/app
$ cd $GOPATH/src/github.com/convox/app
$ make test
$ make build

$ cat fixtures/web_postgis.yml | docker run -i convox/app
{
  "AWSTemplateFormatVersion": "2010-09-09",
  ...
}
```

## Contributing

* Open a [GitHub Issue](https://github.com/convox/app/issues/new) for bugs and feature requests
* Initiate a [GitHub Pull Request](https://help.github.com/articles/using-pull-requests/) for patches

## See Also

* [convox/app](https://github.com/convox/app)
* [convox/build](https://github.com/convox/build)
* [convox/cli](https://github.com/convox/cli)
* [convox/kernel](https://github.com/convox/kernel)

## License

Apache 2.0 &copy; 2015 Convox, Inc.
