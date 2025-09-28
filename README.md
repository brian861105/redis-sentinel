# Redis Sentinel 高可用性配置

本專案展示如何使用 Docker 配置 Redis Sentinel 高可用性集群，包含故障檢測、自動故障轉移和服務發現功能。

## 架構概述

- **Redis Master**: 主要的 Redis 實例 (172.19.0.2:6379)
- **Redis Slave1**: 第一個從節點 (172.19.0.3:6379)
- **Redis Slave2**: 第二個從節點 (172.19.0.4:6379)
- **Sentinel Cluster**: 3 個 Sentinel 實例監控 Redis 集群 (26379-26381)
- **Go Client**: 示範應用程式展示 Sentinel 客戶端連接

## 目錄結構

```
redis-sentinel/
├── docker-compose.yml          # Redis Master/Slave 配置
├── redis.sh                    # 取得 Redis IP 地址的工具腳本
├── sentinel/
│   ├── docker-compose.yml      # Sentinel 集群配置
│   ├── sentinel1.conf          # Sentinel 1 配置
│   ├── sentinel2.conf          # Sentinel 2 配置
│   └── sentinel3.conf          # Sentinel 3 配置
├── web/
│   └── main.go                 # Go 客戶端示範程式
└── README.md
```

## 快速開始

### 1. 啟動 Redis 集群

```bash
# 啟動 Redis Master 和 Slaves
docker compose up -d redis-master redis-slave1 redis-slave2

# 取得 Redis 實例的 IP 地址
./redis.sh
```

### 2. 啟動 Sentinel 集群

```bash
# 切換到 sentinel 目錄並啟動
cd sentinel
docker compose up -d

# 回到主目錄
cd ..
```

### 3. 驗證部署

```bash
# 檢查 Sentinel 狀態
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel masters

# 檢查 Sentinel 是否發現其他節點
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel sentinels mymaster

# 檢查 Redis 主從複製狀態
docker exec redis-master redis-cli -h 172.19.0.2 -p 6379 info replication
```

### 4. 啟動示範應用程式

```bash
# 啟動 Go 客戶端
docker compose up -d go-server

# 訪問應用程式
curl http://localhost:8080
```

## Sentinel 配置說明

### 關鍵配置參數

```conf
port 26379                                          # Sentinel 監聽端口
sentinel monitor mymaster 172.19.0.2 6379 1        # 監控配置 (quorum=1)
sentinel down-after-milliseconds mymaster 30000     # 故障檢測時間 (30秒)
sentinel failover-timeout mymaster 180000           # 故障轉移超時 (3分鐘)
sentinel parallel-syncs mymaster 1                  # 並發同步數量
```

### 配置調整說明

- **Quorum=1**: 適合測試環境，只需要 1 個 Sentinel 同意即可執行故障轉移
- **30秒故障檢測**: 增加容錯性，避免因為短暫網絡問題觸發不必要的故障轉移
- **統一端口配置**: 所有 Sentinel 容器內都使用 26379，透過 Docker 端口映射區分

## 故障轉移測試

### 模擬 Master 故障

```bash
# 停止 Redis Master
docker stop redis-master

# 觀察 Sentinel 日誌
docker logs redis-sentinel-1 -f

# 檢查新的 Master
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel masters
```

### 預期行為

1. **故障檢測**: Sentinel 檢測到 Master 離線
2. **故障轉移**: 其中一個 Slave 被提升為新 Master
3. **配置更新**: 其他節點自動重新配置指向新 Master
4. **舊 Master 恢復**: 當原 Master 重新上線時，會自動變成 Slave

## 常見問題排除

### Tilt 模式問題

如果 Sentinel 頻繁進入 tilt 模式：

```bash
# 檢查 tilt 模式日誌
docker logs redis-sentinel-1 | grep "+tilt"
```

**解決方案**:
- 確保所有 Sentinel 容器內使用相同端口 (26379)
- 調整 quorum 為較小值 (測試環境建議為 1)
- 增加 down-after-milliseconds 時間

### 網絡連接問題

```bash
# 檢查容器網絡
docker network inspect redis-sentinel_default

# 測試容器間連接
docker exec redis-sentinel-1 redis-cli -h 172.19.0.2 -p 6379 ping
```

### 配置檔案權限問題

如果遇到 "Could not rename tmp config file" 錯誤：

```bash
# 停止容器後重新啟動
docker compose down
docker compose up -d
```

## 監控指令

### Sentinel 狀態監控

```bash
# 查看 Master 資訊
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel masters

# 查看 Slave 資訊
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel slaves mymaster

# 查看其他 Sentinel
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel sentinels mymaster

# 取得當前 Master 地址
docker exec redis-sentinel-1 redis-cli -p 26379 sentinel get-master-addr-by-name mymaster
```

### Redis 複製狀態

```bash
# 檢查複製資訊
docker exec redis-master redis-cli info replication
docker exec redis-slave1 redis-cli info replication
docker exec redis-slave2 redis-cli info replication
```

## 生產環境建議

1. **調整 Quorum**: 生產環境建議設定為 `(sentinel_count/2) + 1`
2. **分散部署**: Sentinel 應部署在不同的物理節點
3. **監控告警**: 整合監控系統監控 Sentinel 狀態
4. **備份策略**: 定期備份 Redis 資料
5. **網絡隔離**: 使用專用網絡確保 Sentinel 間通信安全

## 相關連結

- [Redis Sentinel 官方文檔](https://redis.io/docs/manual/sentinel/)
- [Docker Compose 配置參考](https://docs.docker.com/compose/)
- [Redis 高可用性最佳實踐](https://redis.io/docs/manual/sentinel/#high-availability-best-practices)