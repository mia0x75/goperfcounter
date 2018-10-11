Example Of gopfc
=====


这里是，gopfc提供的一个完整的例子。

按照如下指令，运行这个例子。

```bash
# install
cd $GOPATH/src/github.com/mia0x75
git clone https://github.com/mia0x75/gopfc.git
cd $GOPATH/src/github.com/mia0x75/gopfc
go get ./...

# run
cd $GOPATH/src/github.com/mia0x75/gopfc/example/scripts
./debug build && ./debug start

# proc
./debug proc metrics/json   # list all metrics in json 
./debug proc metrics/falcon # list all metrics in falcon-model

```
