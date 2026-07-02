-- OpenConstructionERP
-- V035: cost_transactions supports chapter-level entries (VO incorporation,
--       IPC payments, overheads) — boq_item_id becomes optional.
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

ALTER TABLE cost_transactions ALTER COLUMN boq_item_id DROP NOT NULL;
ALTER TABLE cost_transactions ADD CONSTRAINT cost_tx_target
    CHECK (boq_item_id IS NOT NULL OR cbs_chapter_id IS NOT NULL);
