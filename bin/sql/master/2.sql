select dd.id
from bra.doc_d_001 dd
         cross join lateral (
    select pg_sleep(0.02)
where dd.id is not null
    ) s
limit 400;