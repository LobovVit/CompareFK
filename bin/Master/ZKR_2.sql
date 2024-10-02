--Заявки по которым должны быть проводки
--ЗКР
SELECT z.inf_guid
FROM
	dc_MSC_ApplCashFlow z
JOIN doc d ON d.docid = z.docid
JOIN docstate ds ON ds.docstateid = d.docstateid
WHERE
	INF_CREATIONDATE BETWEEN to_date('20.01.2024','DD.MM.YYYY') AND to_date('01.02.2024','DD.MM.YYYY')
	AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')