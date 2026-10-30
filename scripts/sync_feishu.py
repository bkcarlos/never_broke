#!/usr/bin/env python3
"""One-way Base -> NeverBroke sync; conflicts stop publication.
Requires an authorized lark-cli identity and NEVERBROKE_EMAIL/PASSWORD or TOKEN.
Default is preview only; --commit applies only conflict-free batches.
"""
import argparse, concurrent.futures, json, os, subprocess, tempfile
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.parse import urlparse
from curated_bundle import build
BASE=os.getenv('NEVERBROKE_FEISHU_BASE_TOKEN','')
TABLES=json.loads(os.getenv('NEVERBROKE_FEISHU_TABLES','{}'))
def export_table(key,table,folder):
    def cli(args):
        r=subprocess.run(['lark-cli','base',*args,'--base-token',BASE,'--as','user'],cwd=folder,capture_output=True,text=True,check=True)
        out=json.loads(r.stdout)
        if out.get('ok') is False: raise RuntimeError('Base request failed: '+key)
        return out
    fields=cli(['+field-list','--table-id',table])['data']['fields']
    fields=[f for f in fields if not any(w in f.get('name','').lower() for w in ['email','mail','邮件','正文'])]
    rows=[];offset=0;rev=None
    while True:
        target=f'{key}-{offset}.ndjson'
        args=['+record-list','--table-id',table,'--offset',str(offset),'--format','ndjson','--output',target]
        for f in fields: args.extend(['--field-id',f['id']])
        page=cli(args)
        if rev is not None and rev!=page['rev']: raise RuntimeError('Source changed during export; retry required')
        rev=page['rev']
        rows.extend(json.loads(line) for line in (folder/target).read_text().splitlines() if line.strip())
        if not page['has_more']: break
        if page['next_offset']<=offset: raise RuntimeError('Invalid pagination cursor')
        offset=page['next_offset']
    if len(rows)!=len({r['record_id'] for r in rows}): raise RuntimeError('Duplicate source IDs')
    return key,rows,rev

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--commit',action='store_true');parser.add_argument('--url',default=os.getenv('NEVERBROKE_URL','http://127.0.0.1:18080'));parser.add_argument('--output',type=Path)
    args=parser.parse_args();url=args.url.rstrip('/');u=urlparse(url)
    if not BASE or not isinstance(TABLES,dict) or set(TABLES)!={'cmb','ceb','monthly','apple','calendar'} or not all(isinstance(v,str) and v for v in TABLES.values()):
        raise RuntimeError('Configure NEVERBROKE_FEISHU_BASE_TOKEN and NEVERBROKE_FEISHU_TABLES with all five source tables')
    if u.scheme!='https' and not(u.scheme=='http' and u.hostname in ('localhost','127.0.0.1')): raise RuntimeError('Use HTTPS or loopback')
    token=os.getenv('NEVERBROKE_TOKEN','')
    def request(path,body):
        headers={'Content-Type':'application/json'}
        if token: headers['Authorization']='Bearer '+token
        req=Request(url+'/api'+path,data=json.dumps(body,ensure_ascii=False).encode(),headers=headers,method='POST')
        with urlopen(req,timeout=120) as resp: result=json.load(resp)
        if result.get('code',0)!=0: raise RuntimeError(result.get('message','Request failed'))
        return result.get('data',result)
    if not token:
        auth=request('/auth/login',{'email':os.environ['NEVERBROKE_EMAIL'],'password':os.environ['NEVERBROKE_PASSWORD']})
        token=auth['access_token']
    records=0;conflicts=0
    try:
        with tempfile.TemporaryDirectory(prefix='neverbroke-sync-') as work:
            folder=Path(work)
            with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
                results=list(pool.map(lambda kv:export_table(*kv,folder),TABLES.items()))
            bundle=build({k:rows for k,rows,_ in results})
            if args.output: args.output.write_text(json.dumps(bundle,ensure_ascii=False))
            preview=request('/workspace/import/preview',{'bundle':bundle})
            print(json.dumps({k:preview[k] for k in ('new','updated','skipped','conflicts','totals')},ensure_ascii=False))
            records=len(bundle['records']);conflicts=preview['conflicts']
            if conflicts: raise RuntimeError('Conflicts require review; export with --output and review in the app.')
            if args.commit:
                result=request('/workspace/import/commit',{'bundle':bundle,'token':preview['token'],'resolutions':{}})
                print(json.dumps(result,ensure_ascii=False))
        request('/workspace/sync',{'status':'success' if args.commit else 'preview','records':records,'conflicts':0})
    except Exception as exc:
        try:
            request('/workspace/sync',{'status':'conflict' if conflicts else 'failed','records':records,'conflicts':conflicts,'error':str(exc)[:1000]})
        except Exception:
            pass
        raise
if __name__=='__main__': main()
