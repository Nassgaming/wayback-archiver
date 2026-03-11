-- 去掉 content_hash 唯一约束
-- 允许同一内容（相同哈希）对应多个 URL 的资源记录
-- 文件层面仍按哈希去重，不重复存储
ALTER TABLE resources DROP CONSTRAINT resources_content_hash_key;
