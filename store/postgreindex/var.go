package postgreindex

var tsvector = "to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content,''))"
