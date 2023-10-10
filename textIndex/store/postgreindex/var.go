package postgreindex

var tsvector = "to_tsvector('simple', title || ' ' || content)"

// const matchDocCountQuery = `
// SELECT COUNT(*) FROM documents
// WHERE (to_tsvector('simple', title || ' ' || content) @@ plainto_tsquery('simple', $1))
// `

// /* MATCH */
// const matchDocQuery = `
// SELECT linkID, url, title, content, indexed_at, pagerank
// FROM documents
// WHERE (to_tsvector('simple', title || ' ' || content) @@ plainto_tsquery('simple', $1))
// ORDER BY
// 	ts_rank(to_tsvector('simple', title || ' ' || content), plainto_tsquery('simple', $1)) DESC,
// 	pagerank DESC
// OFFSET ($2) ROWS
// FETCH FIRST ($3) ROWS ONLY;
// `

/* MATCH PHRASE
SELECT id, title, content, score
FROM your_table
WHERE to_tsvector('english', title || ' ' || content) @@ to_tsquery('english', 'your_search_query')
AND to_tsvector('english', title || ' ' || content) @@ to_tsquery('english', 'your_search_query') && to_tsquery('english', 'your_search_query')
ORDER BY ts_rank(to_tsvector('english', title || ' ' || content), to_tsquery('english', 'your_search_query')) DESC, score DESC;
*/
