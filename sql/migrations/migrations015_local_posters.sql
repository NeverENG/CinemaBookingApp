-- Keep demo artwork self-hosted so deployments do not depend on external image CDNs.
UPDATE movies
SET cover_url = CASE id
    WHEN 1 THEN '/posters/dune-3.jpg'
    WHEN 2 THEN '/posters/wandering-earth-3.jpg'
    WHEN 3 THEN '/posters/nezha-2.jpg'
    WHEN 4 THEN '/posters/zootopia-2.jpg'
END
WHERE id IN (1, 2, 3, 4);
