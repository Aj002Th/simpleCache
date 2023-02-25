### simpleCache

a groupcache-like cache lib

项目描述：go标准库实现的分布式缓存库

主要功能：针对静态资源进行缓存，支持多节点

### 技术点：
- LRU实现淘汰策略
- 一致性哈希实现负载均衡
- singleflight机制防止缓存击穿
- protocol buffers编码提高传输效率

### 例子
example 文件夹下有单机和多机两个例子, 打开文件夹直接运行对应例子的 run.sh 即可