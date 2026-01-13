# RocksDB 安装和配置指南（Ubuntu）

## 📋 前置要求

- Ubuntu 18.04 或更高版本
- sudo 权限
- 至少 2GB 可用磁盘空间
- 至少 2GB 可用内存

## 🚀 快速安装

### 方法 1：使用自动化脚本（推荐）

```bash
# 1. 进入脚本目录
cd poc-deploy/localnode/scripts

# 2. 添加执行权限
chmod +x install_rocksdb.sh

# 3. 运行安装脚本
./install_rocksdb.sh
```

脚本会自动：
- ✅ 安装所有依赖
- ✅ 克隆 RocksDB v8.9.1
- ✅ 编译共享库
- ✅ 安装到系统
- ✅ 配置 ldconfig
- ✅ 验证安装

### 方法 2：手动安装

#### 步骤 1：安装依赖

```bash
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    pkg-config \
    cmake \
    git \
    zlib1g-dev \
    libbz2-dev \
    libsnappy-dev \
    liblz4-dev \
    libzstd-dev \
    libjemalloc-dev \
    libgflags-dev
```

#### 步骤 2：克隆和编译 RocksDB

```bash
# 克隆 RocksDB
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
git checkout v8.9.1

# 编译（使用所有 CPU 核心）
make clean
CXXFLAGS='-march=native -DNDEBUG' make -j$(nproc) shared_lib

# 安装
sudo make install-shared

# 配置 ldconfig
echo '/usr/local/lib' | sudo tee /etc/ld.so.conf.d/rocksdb.conf
sudo ldconfig
```

#### 步骤 3：验证安装

```bash
# 检查 RocksDB 是否安装成功
ldconfig -p | grep librocksdb

# 应该看到类似输出：
# librocksdb.so.8 (libc6,x86-64) => /usr/local/lib/librocksdb.so.8
# librocksdb.so (libc6,x86-64) => /usr/local/lib/librocksdb.so
```

## 🔧 编译 seid（带 RocksDB 支持）

### 方法 1：使用 Makefile（推荐）

```bash
cd /path/to/sei-chain

# 编译并安装 seid
make install-rocksdb
```

### 方法 2：手动编译

```bash
cd /path/to/sei-chain

# 设置环境变量
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lz -lbz2 -lsnappy -llz4 -lzstd -ljemalloc"

# 编译
go install -tags "rocksdbBackend" ./cmd/seid
```

### 验证编译

```bash
# 检查 seid 是否链接了 RocksDB
ldd $(which seid) | grep rocksdb

# 应该看到：
# librocksdb.so.8 => /usr/local/lib/librocksdb.so.8
```

## ⚙️ 配置 RocksDB

### 修改 app.toml

```bash
vim poc-deploy/localnode/config/app.toml
```

找到 `[state-store]` 部分，修改：

```toml
[state-store]

# Enable defines if the state-store should be enabled for historical queries.
ss-enable = true

# DBBackend defines the backend database used for state-store.
# Supported backends: pebbledb, rocksdb
ss-backend = "rocksdb"  # ← 改为 rocksdb
```

### 重新初始化链

```bash
cd poc-deploy/localnode/scripts

# 清理旧数据
./clean.sh

# 重新初始化
./step1_configure_init.sh
./step2_genesis.sh
./step3_config_override.sh

# 启动节点
./step4_start_sei.sh
```

## 🔍 故障排查

### 问题 1：找不到 librocksdb.so

**错误信息**：
```
error while loading shared libraries: librocksdb.so.8: cannot open shared object file
```

**解决方案**：
```bash
# 重新配置 ldconfig
echo '/usr/local/lib' | sudo tee /etc/ld.so.conf.d/rocksdb.conf
sudo ldconfig

# 验证
ldconfig -p | grep librocksdb
```

### 问题 2：编译 seid 时找不到 RocksDB 头文件

**错误信息**：
```
fatal error: rocksdb/c.h: No such file or directory
```

**解决方案**：
```bash
# 检查头文件是否存在
ls /usr/local/include/rocksdb/

# 如果不存在，重新安装 RocksDB
cd rocksdb
sudo make install-shared
```

### 问题 3：编译时链接错误

**错误信息**：
```
undefined reference to `rocksdb_xxx'
```

**解决方案**：
```bash
# 确保设置了正确的 CGO 标志
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lz -lbz2 -lsnappy -llz4 -lzstd -ljemalloc"

# 重新编译
make install-rocksdb
```

### 问题 4：运行时性能问题

**症状**：RocksDB 性能不如预期

**优化建议**：

1. **增加 block cache**（修改 `sei-db/ss/rocksdb/opts.go`）：
```go
// 从 1GB 增加到 4GB
bbto.SetBlockCache(grocksdb.NewLRUCache(4 << 30))
```

2. **调整压缩级别**：
```go
// 降低压缩级别以提升写入速度
compressOpts.Level = 6  // 从 12 降到 6
```

3. **增加并行度**：
```go
// 增加并行线程数
opts.IncreaseParallelism(runtime.NumCPU() * 2)
```

## 📊 性能对比

| 数据库 | 写入速度 | 读取速度 | 磁盘占用 | 内存占用 | 推荐场景 |
|--------|---------|---------|---------|---------|---------|
| PebbleDB | 快 | 快 | 中 | 低 | 生产环境 |
| RocksDB | 很快 | 很快 | 低 | 中 | 高性能需求 |

## 🎯 最佳实践

1. **生产环境**：推荐使用 PebbleDB（纯 Go，更稳定）
2. **性能测试**：推荐使用 RocksDB（更快，更成熟）
3. **开发环境**：两者都可以

## 📚 参考资料

- [RocksDB 官方文档](https://github.com/facebook/rocksdb/wiki)
- [sei-db RocksDB 实现](../../sei-db/ss/rocksdb/)
- [Cosmos SDK 数据库后端](https://docs.cosmos.network/)

## ❓ 常见问题

**Q: RocksDB 和 PebbleDB 有什么区别？**

A: 
- RocksDB：C++ 实现，性能更好，需要 CGO
- PebbleDB：纯 Go 实现，更简单，推荐生产环境

**Q: 可以在运行中切换数据库吗？**

A: 不可以，需要重新初始化链

**Q: RocksDB 占用多少磁盘空间？**

A: 取决于数据量，通常比 PebbleDB 少 20-30%

**Q: 如何卸载 RocksDB？**

A:
```bash
sudo rm -rf /usr/local/lib/librocksdb*
sudo rm -rf /usr/local/include/rocksdb
sudo rm /etc/ld.so.conf.d/rocksdb.conf
sudo ldconfig
```

