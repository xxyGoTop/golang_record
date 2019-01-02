### cache2go (key/value缓存 并发安全具有过去功能)
>- 学习cache2go库源码
>- 安全操作map
>- 定时器 timer
>- 指针操作

### pool (协程池)
>- 创建一个协程池
>- 新增加任务，判断是否有工作协程，无则新增一个执行，任务处理完后回收，并通知
>- 若有空闲的工作协程，直接取出去处理任务，无则新增，直到最大容量，超过最大容量就等待worker的释放
>- 多协程操作需加锁
>- 原子操作 确保执行不可中断