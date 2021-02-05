第9章 统计系统之统计逻辑
最贴近业务需求的系统子模块，其中采用HyperLogLog能力实现UV天级去重，定义了用于存储数据所需的结构体后，对过往数据进行架构封装，投递至待存储通道，待存储器消费。

共 2 节 (30分钟) 收起列表

 9-1 统计分析模块PVUV统计（上） (21:35)
 9-2 统计分析模块PVUV统计（下） (08:23)



Redis　Clinet
```
	"github.com/mediocregopher/radix.v2/redis"


  var redisClient redis.Client


  redisClient, err:= redis.Dial("tcp", "localhost:6379")

  if err != nil {
    log.Fatalln(...)
  } else {
    defer redisCli.Close()
  }

```


并发模式下不要用 RedisClient， 而应该用 RedisPool, 把 RedisPool传入routine func

```
  "github.com/mediocregopher/radix.v2/pool"
  
  // N client
  redisPool, err := pool.New( "tcp", "192.168.1.100:6379", N );

	if err != nil {
		log.Fatalln("Redis pool created failed")
		panic(err)
	} else {
		go func() {
			for {
				redisPool.Cmd("PING")
				time.Sleep(3 * time.Second)
			}
		}()
	}
```