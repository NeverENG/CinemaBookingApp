-- 大盘演示种子数据（migrations010，入口在引导 demo 用户后单独执行）

-- ============ 基础数据 ============
INSERT INTO cinemas (id, name, city, address, longitude, latitude, status)
VALUES (2, '环球影城', '北京', '朝阳区演示路 2 号', 116.40, 39.90, 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO movies (id, title, cover_url, trailer_url, description, duration_minutes, genre, release_date, rating, status) VALUES
    (2, '流浪地球3', 'https://picsum.photos/seed/movie2/400/600', '', '演示影片', 173, '科幻', '2026-07-15', 9.1, 'ON_SALE'),
    (3, '哪吒2', 'https://picsum.photos/seed/movie3/400/600', '', '演示影片', 144, '动画', '2026-08-10', 9.4, 'ON_SALE'),
    (4, '疯狂动物城2', 'https://picsum.photos/seed/movie4/400/600', '', '演示影片', 128, '动画', '2026-08-20', 8.6, 'ON_SALE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO halls (id, cinema_id, name, seat_layout_json, status) VALUES
    (100, 1, '2号厅', '{"rows":2,"cols":3}', 'ACTIVE'),
    (200, 2, 'IMAX厅', '{"rows":3,"cols":3}', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO seats (id, hall_id, row_no, col_no, seat_no, type, status) VALUES
    (1001, 100, 1, 1, 'A1', 'STANDARD', 'ENABLED'),
    (1002, 100, 1, 2, 'A2', 'STANDARD', 'ENABLED'),
    (1003, 100, 1, 3, 'A3', 'STANDARD', 'ENABLED'),
    (1004, 100, 2, 1, 'B1', 'STANDARD', 'ENABLED'),
    (1005, 100, 2, 2, 'B2', 'STANDARD', 'ENABLED'),
    (1006, 100, 2, 3, 'B3', 'STANDARD', 'ENABLED'),
    (2001, 200, 1, 1, 'A1', 'VIP', 'ENABLED'),
    (2002, 200, 1, 2, 'A2', 'VIP', 'ENABLED'),
    (2003, 200, 1, 3, 'A3', 'VIP', 'ENABLED'),
    (2004, 200, 2, 1, 'B1', 'STANDARD', 'ENABLED'),
    (2005, 200, 2, 2, 'B2', 'STANDARD', 'ENABLED'),
    (2006, 200, 2, 3, 'B3', 'STANDARD', 'ENABLED'),
    (2007, 200, 3, 1, 'C1', 'STANDARD', 'ENABLED'),
    (2008, 200, 3, 2, 'C2', 'STANDARD', 'ENABLED'),
    (2009, 200, 3, 3, 'C3', 'STANDARD', 'ENABLED')
ON CONFLICT DO NOTHING;

-- 历史场次（CLOSED，用于票房）
INSERT INTO show_sessions (id, cinema_id, hall_id, movie_id, start_time, end_time, base_price_cents, status) VALUES
    (101, 1, 1, 1, date_trunc('day', now()) - interval '14 days' + interval '19 hours', date_trunc('day', now()) - interval '14 days' + interval '21 hours', 5000, 'CLOSED'),
    (102, 1, 100, 2, date_trunc('day', now()) - interval '12 days' + interval '19 hours', date_trunc('day', now()) - interval '12 days' + interval '21 hours', 6000, 'CLOSED'),
    (103, 2, 200, 3, date_trunc('day', now()) - interval '10 days' + interval '19 hours', date_trunc('day', now()) - interval '10 days' + interval '21 hours', 4500, 'CLOSED'),
    (104, 1, 1, 4, date_trunc('day', now()) - interval '8 days' + interval '19 hours', date_trunc('day', now()) - interval '8 days' + interval '21 hours', 5500, 'CLOSED'),
    (105, 2, 200, 1, date_trunc('day', now()) - interval '7 days' + interval '19 hours', date_trunc('day', now()) - interval '7 days' + interval '21 hours', 5000, 'CLOSED'),
    (106, 1, 100, 2, date_trunc('day', now()) - interval '6 days' + interval '19 hours', date_trunc('day', now()) - interval '6 days' + interval '21 hours', 6000, 'CLOSED'),
    (107, 1, 1, 3, date_trunc('day', now()) - interval '5 days' + interval '19 hours', date_trunc('day', now()) - interval '5 days' + interval '21 hours', 4500, 'CLOSED'),
    (108, 2, 200, 4, date_trunc('day', now()) - interval '3 days' + interval '19 hours', date_trunc('day', now()) - interval '3 days' + interval '21 hours', 5500, 'CLOSED'),
    (109, 1, 100, 1, date_trunc('day', now()) - interval '2 days' + interval '19 hours', date_trunc('day', now()) - interval '2 days' + interval '21 hours', 5000, 'CLOSED'),
    (110, 1, 1, 2, date_trunc('day', now()) - interval '1 day' + interval '19 hours', date_trunc('day', now()) - interval '1 day' + interval '21 hours', 6000, 'CLOSED')
ON CONFLICT (id) DO UPDATE SET
    cinema_id = EXCLUDED.cinema_id,
    hall_id = EXCLUDED.hall_id,
    movie_id = EXCLUDED.movie_id,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    base_price_cents = EXCLUDED.base_price_cents,
    status = EXCLUDED.status;

-- 未开场次（OPEN，可买票）
INSERT INTO show_sessions (id, cinema_id, hall_id, movie_id, start_time, end_time, base_price_cents, status) VALUES
    (201, 1, 1, 1, date_trunc('day', now()) + interval '1 day' + interval '19 hours', date_trunc('day', now()) + interval '1 day' + interval '21 hours', 5000, 'OPEN'),
    (202, 2, 200, 3, date_trunc('day', now()) + interval '2 days' + interval '19 hours', date_trunc('day', now()) + interval '2 days' + interval '21 hours', 4500, 'OPEN'),
    (203, 1, 100, 1, date_trunc('day', now()) + interval '3 days' + interval '19 hours', date_trunc('day', now()) + interval '3 days' + interval '21 hours', 5500, 'OPEN')
ON CONFLICT (id) DO UPDATE SET
    cinema_id = EXCLUDED.cinema_id,
    hall_id = EXCLUDED.hall_id,
    movie_id = EXCLUDED.movie_id,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    base_price_cents = EXCLUDED.base_price_cents,
    status = EXCLUDED.status;

-- ============ 幂等清理（只清种子数据） ============
DELETE FROM seat_locks WHERE order_no LIKE 'OSEED%';
DELETE FROM box_office_ledger WHERE biz_no LIKE 'OSEED%' OR biz_no LIKE 'RFSEED%';
DELETE FROM order_items WHERE order_no LIKE 'OSEED%';
DELETE FROM payment_transactions WHERE biz_no LIKE 'OSEED%';
DELETE FROM refunds WHERE refund_no LIKE 'RFSEED%';
DELETE FROM orders WHERE order_no LIKE 'OSEED%';

-- ============ 已支付订单（每个历史场次 1~2 单） ============
INSERT INTO orders (order_no, user_id, session_id, cinema_id, movie_id, status, total_cents, discount_cents, coupon_cents, paid_cents, expire_at, version, created_at, paid_at)
SELECT 'OSEED' || s.id || '_1',
       (SELECT id FROM users WHERE username = 'demo'),
       s.id, s.cinema_id, s.movie_id, 'PAID',
       s.base_price_cents, 0, 0, s.base_price_cents,
       s.start_time - interval '1 day', 1,
       s.start_time - interval '2 hours', s.start_time - interval '90 minutes'
FROM show_sessions s WHERE s.id BETWEEN 101 AND 110;

INSERT INTO orders (order_no, user_id, session_id, cinema_id, movie_id, status, total_cents, discount_cents, coupon_cents, paid_cents, expire_at, version, created_at, paid_at)
SELECT 'OSEED' || s.id || '_2',
       (SELECT id FROM users WHERE username = 'demo'),
       s.id, s.cinema_id, s.movie_id, 'PAID',
       s.base_price_cents, 0, 0, s.base_price_cents,
       s.start_time - interval '1 day', 1,
       s.start_time - interval '90 minutes', s.start_time - interval '30 minutes'
FROM show_sessions s WHERE s.id BETWEEN 101 AND 110 AND s.id % 2 = 0;

-- 订单明细（每单 1 张票）
INSERT INTO order_items (order_no, session_id, seat_id, seat_no, price_cents, ticket_no, created_at)
SELECT o.order_no, o.session_id,
       (SELECT id FROM seats WHERE hall_id = s.hall_id ORDER BY id LIMIT 1 OFFSET CASE WHEN o.order_no LIKE '%_2' THEN 1 ELSE 0 END),
       (SELECT seat_no FROM seats WHERE hall_id = s.hall_id ORDER BY id LIMIT 1 OFFSET CASE WHEN o.order_no LIKE '%_2' THEN 1 ELSE 0 END),
       o.paid_cents, 'TKSEED' || o.order_no, o.paid_at
FROM orders o JOIN show_sessions s ON s.id = o.session_id
WHERE o.order_no LIKE 'OSEED%';

-- 支付交易
INSERT INTO payment_transactions (transaction_no, biz_type, biz_no, user_id, amount_cents, channel, external_trade_no, status, version, created_at, paid_at)
SELECT 'T' || o.order_no, 'ORDER_PAY', o.order_no, o.user_id, o.paid_cents, 'MOCK_PAY', 'EXTSEED' || o.order_no, 'SUCCESS', 1, o.paid_at, o.paid_at
FROM orders o WHERE o.order_no LIKE 'OSEED%';

-- 座位锁（BOOKED）
INSERT INTO seat_locks (session_id, seat_id, user_id, order_no, lock_token, status, expires_at, created_at)
SELECT s.id, i.seat_id, o.user_id, o.order_no, 'LOCKSEED' || o.order_no, 'BOOKED', s.start_time, o.paid_at
FROM orders o
JOIN show_sessions s ON s.id = o.session_id
JOIN order_items i ON i.order_no = o.order_no
WHERE o.order_no LIKE 'OSEED%';

-- 票房事件（ORDER_PAID）
INSERT INTO box_office_ledger (biz_type, biz_no, stat_date, cinema_id, movie_id, order_delta, ticket_delta, gross_delta, refund_delta, created_at)
SELECT 'ORDER_PAID', o.order_no, o.paid_at::date, o.cinema_id, o.movie_id, 1, 1, o.paid_cents, 0, o.paid_at
FROM orders o WHERE o.order_no LIKE 'OSEED%';

-- ============ 一笔退款（演示退款曲线） ============
UPDATE orders SET status = 'REFUNDED'
WHERE order_no = 'OSEED103_1';

UPDATE payment_transactions SET status = 'REFUNDED'
WHERE biz_no = 'OSEED103_1';

INSERT INTO refunds (refund_no, order_no, user_id, amount_cents, reason, status, external_refund_no, refunded_at, created_at)
SELECT 'RFSEED103', order_no, user_id, paid_cents, 'seed_demo', 'SUCCESS', 'EXT-RFSEED103', paid_at + interval '10 minutes', paid_at + interval '10 minutes'
FROM orders WHERE order_no = 'OSEED103_1';

INSERT INTO box_office_ledger (biz_type, biz_no, stat_date, cinema_id, movie_id, order_delta, ticket_delta, gross_delta, refund_delta, created_at)
SELECT 'ORDER_REFUND', refund_no, r.refunded_at::date, o.cinema_id, o.movie_id, -1, -1, 0, r.amount_cents, r.refunded_at
FROM refunds r JOIN orders o ON o.order_no = r.order_no
WHERE r.refund_no = 'RFSEED103';

-- 退款单的锁释放
UPDATE seat_locks SET status = 'RELEASED', released_at = now()
WHERE order_no = 'OSEED103_1';

-- ============ 重建日聚合（由 ledger 全量重算） ============
DELETE FROM daily_box_office;
INSERT INTO daily_box_office (stat_date, cinema_id, movie_id, order_count, ticket_count, gross_cents, refund_cents, net_cents, updated_at)
SELECT stat_date, cinema_id, movie_id,
       SUM(order_delta), SUM(ticket_delta), SUM(gross_delta), SUM(refund_delta),
       SUM(gross_delta) - SUM(refund_delta), now()
FROM box_office_ledger
GROUP BY stat_date, cinema_id, movie_id;
