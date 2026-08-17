# rm-search 私有部署

独立的 rm-search 全套栈: MySQL (原始数据) + Meilisearch (索引) + rm-search (服务与自动增量) + nginx (对外暴露)。适用于给其他服务 (如 aim-feishu-rm-assistant) 提供独立的数据源。

## 组成

| 服务 | 用途 | 端口 |
| --- | --- | --- |
| mysql:8.4 | 原始爬取数据 (bbs_post / announce / attachment) | 容器内 3306 |
| meilisearch:v1.12 | 倒排索引 | 容器内 7700 |
| rm-search | HTTP 服务 + 定时任务 (论坛增量每分钟、公告增量每分钟、词云每 5 分钟) | 容器内 8080 |
| nginx | 对外统一入口, 把 `/api/ms/*` 映射到 rm-search 的 `/ms/*` (与生产路径结构一致) | 宿主机 `LISTEN_PORT` (默认 8081) |

## 部署步骤

```bash
cd deploy/

# 1. 准备配置
cp .env.example .env
cp config.template.yaml config.yaml
#    生成随机密钥并同步填入两处:
#    .env 的 MYSQL_ROOT_PASSWORD / MEILI_MASTER_KEY
#    config.yaml 的 DataSource 密码 / MeiliSearch.APIKey
openssl rand -hex 24   # 可用来生成密钥

# 2. 启动 (首次启动 mysql 会自动执行 database/rm_search.sql 建表)
docker compose up -d --build

# 3. 首次建索引设置 (分词、排序规则)
docker compose run --rm rm-search /usr/local/bin/setup-index

# 4. 全量回填 (后台运行, 论坛帖子 + 公告)
#    全量约需数天; 可先爬近期区间快速可用, 剩余的继续后台爬:
docker compose run --rm rm-search /usr/local/bin/crawl \
    --announce-start 1 --announce-end 3000
nohup docker compose run --rm rm-search /usr/local/bin/crawl \
    --posts-start 0 --posts-end 2000000 --posts-goroutines 50 \
    >> crawl.log 2>&1 &

# 5. 回填完成后把已入库的数据灌进 Meilisearch
#    (增量任务会自动索引之后的新数据; recreate-index 是全量重建)
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
```

## 接入机器人

aim-feishu-rm-assistant 的 `.env` 中:

```
RMSEARCH_BASE_URL=http://<本机地址>:8081
```

## 运维说明

- **自动更新**: rm-search 进程内置定时任务, 论坛三类帖和公告都是每分钟增量同步, 无需外部 CronJob
- **数据目录**: `deploy/data/` (mysql、meili), 备份拷走即可
- **资源参考**: 全量数据下 MySQL 约 2~5 GB, Meilisearch 内存 1~2 GB, 建议 2C4G 以上
- **DJIMetaKey**: 目前论坛接口无需登录, 留空即可; 若日后论坛重新强制登录, 在 config.yaml 填入登录 Cookie 的 `_meta_key`
- **升级**: `git pull && docker compose up -d --build` (数据保留)
