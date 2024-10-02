--Исполенные заявки с проводками
--select e.doc_external_id from acc_entry e where e.doc_external_id = any($1) group by doc_external_id having nvl(sum(amount),0) = 0
select
		doc_external_id	
from acc_transaction trx
where 
	trx.accounting_book_id = 6004
	and trx.doc_external_id = any($1)
	and trx.settlement_date >  TO_DATE('2024-01-01','YYYY-MM-DD')
group by doc_external_id
having sum(nvl(debit_amount,0)) != 0;