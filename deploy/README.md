# rm-search 私有部署

独立的 rm-search 全套栈: PostgreSQL (原始数据) + Meilisearch (索引) + rm-search (服务与自动增量), rm-search 直接对外暴露 (原生支持 `/api/ms/*` 与 `/ms/*` 两种路径前缀)。适用于给其他服务 (如 aim-feishu-rm-assistant) 提供独立的数据源。

## 组成

| 服务 | 用途 | 端口 |
| --- | --- | --- |
| postgres:16 | 原始爬取数据 (bbs_post / announce / attachment) | 容器内 5432 |
| meilisearch:v1.12 | 倒排索引 | 容器内 7700 |
| rm-search | HTTP 服务 + 定时任务 (论坛/公告增量每分钟、追新每天 0/6/12/18 点、历史回填常驻循环、词云每 5 分钟); 对外入口, 原生支持 `/api/ms/*` (与生产路径一致) 和 `/ms/*` | 宿主机 `LISTEN_PORT` (默认 8081) → 容器 8080 |

## 部署步骤

```bash
cd deploy/

# 1. 准备配置
cp .env.example .env
cp config.template.yaml config.yaml
#    生成随机密钥并同步填入两处:
#    .env 的 POSTGRES_PASSWORD / MEILI_MASTER_KEY
#    config.yaml 的 DataSource 密码 / MeiliSearch.APIKey
openssl rand -hex 24   # 可用来生成密钥

# 2. 启动 (rm-search 二进制内嵌建表 DDL, 启动时自动应用, 无需手动初始化)
docker compose up -d --build

# 3. (可选) 加速全量回填
#    rm-search 启动后自动持续回填历史帖子 (后台循环, 每块约 10 万个 ID,
#    RM_SEARCH_CRAWL_CHUNK 可调; 并发默认 20, RM_SEARCH_CRAWL_GOROUTINES 可调;
#    对论坛的全局请求速率默认 40 QPS (RM_SEARCH_BBS_QPS), 收到限流 (405) 时
#    自动全局冷却 30 秒起步、指数退避到最长 10 分钟),
#    触底后标记完成并永久停止;
#    每天 0/6/12/18 点另有追新任务补齐最新帖子。等不及可用 crawl 手动加速:
docker compose run --rm rm-search /usr/local/bin/crawl \
    --announce-start 1 --announce-end 3000
nohup docker compose run --rm rm-search /usr/local/bin/crawl \
    --posts-start 0 --posts-end 2000000 --posts-goroutines 50 \
    >> crawl.log 2>&1 &

# 4. 回填完成后把已入库的历史数据灌进 Meilisearch
#    (增量任务会自动索引新数据, 但不会索引回填的历史帖; recreate-index 是全量重建)
#    回填完成标志: 日志出现 "crawl backfill reached floor ... historical crawling done"
docker compose run --rm rm-search /usr/local/bin/recreate-index
```

crawl 支持断点续爬: 已成功持久化的帖子会跳过 (`--posts-*`), 中断后重新运行同一命令即可从断点继续。公告区间用 `--announce-start/--announce-end` 控制; 当前公告 ID 在 2000 左右, 爬 1~3000 即可覆盖全部。

## 验证

```bash
# 服务健康
curl -s http://localhost:8081/healthz

# 搜索接口 (机器人使用的同款路径)
curl -s -X POST http://localhost:8081/api/ms/indexes/rm-search/search \
  -H "Content-Type: application/json" \
  -d '{"q":"","sort":["create_time:desc"],"limit":3,"attributesToRetrieve":["title","url","create_time"]}'

# 增量任务日志 (每分钟应出现 article/faq/wiki/announce 的检查日志)
docker compose logs -f rm-search | grep -i latest

# 回填进度 (每块约 10 万个 ID, 触底后停止)
docker compose logs -f rm-search | grep "backfill"
```

## 接入机器人

aim-feishu-rm-assistant 的 `.env` 中:

```env
RMSEARCH_BASE_URL=http://<本机地址>:8081
```

## 运维说明

- **自动更新**: rm-search 进程内置定时任务, 论坛三类帖和公告每分钟增量同步, 历史回填常驻运行, 无需外部 CronJob
- **数据目录**: `deploy/data/` (pg、meili), 备份拷走即可
- **资源参考**: 全量数据下 PostgreSQL 约 2~5 GB, Meilisearch 内存 1~2 GB, 建议 2C4G 以上
- **DJIMetaKey**: 目前论坛接口无需登录, 留空即可; 若日后论坛重新强制登录, 在 config.yaml 填入登录 Cookie 的 `_meta_key`
- **升级**: `git pull && docker compose pull rm-search && docker compose up -d` (数据保留)
