-- 演示种子数据（migrations008）

INSERT INTO cinemas (id, name, city, address, longitude, latitude, status)
VALUES (1, '星海影城', '上海', '浦东新区测试路 1 号', 121.47, 31.23, 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO movies (id, title, cover_url, trailer_url, description, duration_minutes, genre, release_date, rating, status)
VALUES (1, '沙丘3', 'https://picsum.photos/seed/movie1/400/600', '', '演示影片', 166, '科幻', '2026-08-01', 8.8, 'ON_SALE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO halls (id, cinema_id, name, seat_layout_json, status)
VALUES (1, 1, '1号厅', '{"rows":2,"cols":3}', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO seats (hall_id, row_no, col_no, seat_no, type, status) VALUES
    (1, 1, 1, 'A1', 'STANDARD', 'ENABLED'),
    (1, 1, 2, 'A2', 'STANDARD', 'ENABLED'),
    (1, 1, 3, 'A3', 'STANDARD', 'ENABLED'),
    (1, 2, 1, 'B1', 'STANDARD', 'ENABLED'),
    (1, 2, 2, 'B2', 'STANDARD', 'ENABLED'),
    (1, 2, 3, 'B3', 'STANDARD', 'ENABLED')
ON CONFLICT (hall_id, row_no, col_no) DO NOTHING;

INSERT INTO show_sessions (id, cinema_id, hall_id, movie_id, start_time, end_time, base_price_cents, status)
VALUES (1, 1, 1, 1, now() + interval '1 day', now() + interval '1 day' + interval '166 minutes', 5000, 'OPEN')
ON CONFLICT (id) DO NOTHING;
