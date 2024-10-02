--ПЗВ
SELECT z.zv_inf_guid
FROM 
	dc_MSC_ApplRefund z
JOIN doc d ON d.docid = z.docid
JOIN docstate ds ON ds.docstateid = d.docstateid
WHERE 
	ZV_INF_CREATIONDATE BETWEEN to_date('01.01.2024','DD.MM.YYYY') AND to_date('20.01.2024','DD.MM.YYYY')
    AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')