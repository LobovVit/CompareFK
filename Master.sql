--ЗКР
SELECT	z.inf_guid--,	'ЗКР' res
FROM
    dc_MSC_ApplCashFlow z
        JOIN doc d ON d.docid = z.docid
        JOIN docstate ds ON	ds.docstateid = d.docstateid
WHERE
    d.CREATEDATE BETWEEN to_date('01.01.2023','DD.MM.YYYY') AND to_date('10.09.2023','DD.MM.YYYY')
--  AND z.ZR_SRVC_TOFK_CODE IN ('1500', '7100','5400','4000','0600',
--                              '0700','2300','4300','9100','2800')
  AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')
UNION all
--ЗКC
SELECT z.inf_guid--, 'ЗКC' res
FROM dc_MSC_ApplCashFlowShrt z
         JOIN doc d ON d.docid = z.docid
         JOIN docstate ds ON ds.docstateid = d.docstateid
WHERE d.CREATEDATE BETWEEN to_date('01.01.2023','DD.MM.YYYY') AND to_date('10.09.2023','DD.MM.YYYY')
--  AND z.ZR_SRVC_TOFK_CODE IN ('1500', '7100','5400','4000','0600',
--                              '0700','2300','4300','9100','2800')
  AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')
UNION all
--ЗСВ
SELECT z.inf_guid--, 'ЗСВ' res
FROM dc_MSC_SumApplCashFlowTax z
         JOIN doc d ON d.docid = z.docid
         JOIN docstate ds ON ds.docstateid = d.docstateid
WHERE d.CREATEDATE BETWEEN to_date('01.01.2023','DD.MM.YYYY') AND to_date('10.09.2023','DD.MM.YYYY')
--  AND z.Z_SRVC_TOFK_CODE IN ('1500', '7100','5400','4000','0600',
  --                            '0700','2300','4300','9100','2800')
  AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')
UNION all
--ПЗВ
SELECT z.zv_inf_guid-- 'ПЗВ' res
FROM dc_MSC_ApplRefund z
         JOIN doc d ON d.docid = z.docid
         JOIN docstate ds ON ds.docstateid = d.docstateid
WHERE d.CREATEDATE BETWEEN to_date('01.01.2023','DD.MM.YYYY') AND to_date('10.09.2023','DD.MM.YYYY')
--  AND z.ZV_SRVC_TOFK_CODE IN ('1500', '7100','5400','4000','0600',
  --                             '0700','2300','4300','9100','2800')
  AND ds.systemname in ('Executed', 'CheckedFK', 'REGISTRED')
UNION all
SELECT dd.GLOBALDOCID FROM doc dd where rownum < 20000000