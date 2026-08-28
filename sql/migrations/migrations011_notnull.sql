-- 唯一业务号 NOT NULL（migrations011）

UPDATE order_items SET ticket_no = 'TKBACKFILL' || id WHERE ticket_no IS NULL;
ALTER TABLE order_items ALTER COLUMN ticket_no SET NOT NULL;

UPDATE payment_transactions SET external_trade_no = 'EXTBACKFILL' || transaction_no WHERE external_trade_no IS NULL;
ALTER TABLE payment_transactions ALTER COLUMN external_trade_no SET NOT NULL;

UPDATE refunds SET external_refund_no = 'EXRBACKFILL' || id WHERE external_refund_no IS NULL;
ALTER TABLE refunds ALTER COLUMN external_refund_no SET NOT NULL;
