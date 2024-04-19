### tiny-groupcache

a groupcache-like cache lib

项目描述：go标准库实现的分布式缓存库, 是开源分布式缓存库 groupcache 的一个简要实现

主要功能：针对静态资源进行缓存，支持多节点

### 技术点：
- 使用 LRU 算法实现内存淘汰策略
- 一致性哈希实现负载均衡
- singleflight 机制防止缓存击穿
- protocol buffers 编码提高传输效率

### 例子
example 文件夹下有单机和多机两个例子, 打开文件夹直接运行对应例子的 run.sh 即可