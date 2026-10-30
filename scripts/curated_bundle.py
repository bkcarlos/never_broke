"""Convert fully exported, reviewed Base tables into a versioned import bundle.
No credentials or message content are retained. Monetary arithmetic uses Decimal.
"""
import decimal
D=decimal.Decimal
def build(src):
    ids={r['record_id']:f'{k}:{r["record_id"]}' for k,rows in src.items() for r in rows}
    records=[];accounts={}
    def cents(v):return int((D(str(v))*100).quantize(D('1')))
    def sel(r,k,default=''):return (r.get(k) or [default])[0]
    def links(r,names):return list(dict.fromkeys(ids[v['id']] for n in names for v in (r.get(n) or []) if v['id'] in ids))
    for k,rows in src.items():
     for r in rows:
      key=ids[r['record_id']];cur='CNY' if k in ['cmb','monthly'] else sel(r,'币种','CNY')
      out={'key':key,'currency':cur,'metadata':{n:v for n,v in r.items() if n!='record_id' and not any(x in n.lower() for x in ['email','mail','邮件','正文'])}}
      if k in ['cmb','ceb']:
       typ=sel(r,'交易类型');out.update(kind={'消费':'expense','退款':'refund'}.get(typ,'reference'),date=(r.get('分析日期') or r.get('交易日期') or r.get('记账日期') or r['账单月份']+'-01')[:10],description=r['商户/摘要'],category=sel(r,'消费分类'),tags=r.get('标签') or [])
       amt=r.get('金额' if k=='cmb' else '原币金额');out['amountCents']=abs(cents(amt)) if amt is not None else None
       if out['kind'] in ['expense','refund']:
        card=r['卡号尾号'];ak=f'{k}:{card}:{cur}';out['accountKey']=ak;accounts[ak]={'key':ak,'institution':k,'cardLast4':card,'currency':cur}
       out['note']='；'.join(str(r[n]) for n in ['商品/服务','退款关联说明','日期说明','核对说明'] if r.get(n))
       out['relationStatus']=sel(r,'退款关联状态')
       out['relatedKeys']=links(r,['对应消费']) if typ=='退款' else links(r,['关联退款','Apple收据'])
      elif k=='apple':
       out.update(kind='receipt',date=r['凭证日期'][:10],description=r['商品/服务'],amountCents=cents(r['凭证金额']),category=sel(r,'消费分类'),tags=r.get('标签') or [],relatedKeys=links(r,['关联招商消费','关联光大消费']),relationStatus=sel(r,'匹配状态'),note=r.get('匹配说明') or '')
      elif k=='monthly':
       out.update(kind='statement',date=r['账单月份']+'-01',description=r['账单名称'],note=r.get('核对说明') or '',relatedKeys=links(r,['光大交易明细']))
       if r.get('本期应还') is not None:out['amountCents']=cents(r['本期应还'])
      else:
       kind='daily_summary' if r['事项类型']==['每日消费'] else 'schedule';out.update(kind=kind,date=r['日期'][:10],description=r['事项'],note=r.get('说明') or '',relationStatus=sel(r,'状态'),relatedKeys=links(r,['招商交易','光大交易','关联月度账单']))
       amt=r.get('消费净额') if kind=='daily_summary' else r.get('正式账单金额') if r.get('正式账单金额') is not None else r.get('预计还款金额')
       if amt is not None:out['amountCents']=cents(amt)
      records.append({n:v for n,v in out.items() if v is not None})
    bundle={'version':1,'accounts':list(accounts.values()),'records':records}
    return bundle
