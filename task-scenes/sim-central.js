const fs=require('fs');
const ALVO='/home/user/task-scenes/central-wca/central-wca-1.56.html';
const html=fs.readFileSync(ALVO,'utf8');
let js=html.split('<script>')[1].split('</script>')[0].split("catch(e){ }").join("catch(e){ global.__err.push(String(e&&e.message)); }");
let falhas=0;
function ok(nome,cond,extra){ if(cond) console.log('PASS '+nome); else { falhas++; console.log('FAIL '+nome+(extra?' :: '+extra:'')); } }
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
function makeEl(tag){
  const el={tagName:(tag||'div').toUpperCase(),_isFrag:String(tag).toLowerCase()==='fragment',children:[],_html:'',_text:'',style:{},dataset:{},_cls:new Set(),_listeners:{},offsetLeft:10,offsetWidth:50,_q:{},_parent:null,content:null,src:'',_type:'',
    classList:{ add(...cs){ cs.forEach(c=>el._cls.add(c)); }, remove(...cs){ cs.forEach(c=>el._cls.delete(c)); },
      toggle(c,f){ if(f===undefined){ el._cls.has(c)?el._cls.delete(c):el._cls.add(c); } else if(f){ el._cls.add(c); } else { el._cls.delete(c); } },
      contains(c){ return el._cls.has(c); } },
    appendChild(c){ if(c._isFrag){ c.children.slice().forEach(x=>{ x._parent=el; el.children.push(x); }); c.children=[]; return c; } c._parent=el; el.children.push(c); return c; },
    remove(){ if(el._parent) el._parent.children=el._parent.children.filter(x=>x!==el); },
    addEventListener(t,f){ (el._listeners[t]=el._listeners[t]||[]).push(f); },
    removeAttribute(){}, setAttribute(){}, getAttribute(k){ if(k==='data-val') return el._dv||null; if(k==='data-key') return el._dk||null; return null; },
    closest(sel){ let cur=el; while(cur){ const parts=String(sel).split(','); for(let pt of parts){ pt=pt.trim(); if(pt[0]==='#'&&cur._id===pt.slice(1)) return cur; if(pt[0]==='.'&&cur._cls.has(pt.slice(1))) return cur; } cur=cur._parent; } return null; },
    querySelector(sel){ if(sel==='input[type="checkbox"]'){ const f=el.children.find(c=>String(c.tagName)==='INPUT'&&c._type==='checkbox'); if(f) return f; } if(!el._q[sel]){ const c=makeEl('div'); if(sel[0]==='.') c._cls.add(sel.slice(1)); el._q[sel]=c; } return el._q[sel]; },
    querySelectorAll(sel){ const out=[]; const ehCheckbox=(sel==='input[type="checkbox"]'); const cls=sel[0]==='.'?sel.slice(1):null;
      (function walk(n){ n.children.forEach(c=>{ if(ehCheckbox&&String(c.tagName)==='INPUT'&&c._type==='checkbox') out.push(c); else if(cls&&c._cls.has(cls)) out.push(c); walk(c); }); })(el); return out; },
    cloneNode(deep){ const c=makeEl(el.tagName.toLowerCase()); c._q=el._q; c.content=el.content; c._cls=new Set(el._cls); return c; } };
  Object.defineProperty(el,'className',{ get(){return [...el._cls].join(' ');}, set(v){ el._cls=new Set(String(v).split(/\s+/).filter(Boolean)); }});
  Object.defineProperty(el,'innerHTML',{ get(){return el._html;}, set(v){ el._html=String(v); el.children=[]; }});
  Object.defineProperty(el,'textContent',{ get(){return el._text;}, set(v){ el._text=String(v); }});
  Object.defineProperty(el,'value',{ get(){return el._value!==undefined?el._value:'';}, set(v){ el._value=String(v); }});
  Object.defineProperty(el,'type',{ get(){return el._type;}, set(v){ el._type=String(v); }});
  Object.defineProperty(el,'checked',{ get(){return !!el._checked;}, set(v){ el._checked=!!v; }});
  return el;
}
const registry={}; const createdImgs=[]; const domListeners={};
global.__err=[];
function toggleDe(key, vals, def){
  const tg=makeEl('div'); tg._cls.add('dev-toggle'); tg._dk=key; tg.getAttribute=()=>key;
  tg._opts=[];
  vals.forEach(v=>{ const o=makeEl('button'); o._cls.add('dev-opt'); o._dv=v; if(v===def) o._cls.add('active'); tg.appendChild(o); tg._opts.push(o); });
  tg._opts.forEach(o=>{ o._parent=tg; });
  return tg;
}
function criaModalDev(){
  const modal=makeEl('div'); modal._cls.add('modal-overlay'); registry['devSettingsModal']=modal;
  const toggles=[];
  const spec={ auto_mark_read:{opts:['true','false'],def:'false'},
               auto_reject_call:{opts:['true','false'],def:'false'},
               auto_download_media:{opts:['true','false'],def:'true'},
               eco:{opts:['true','false'],def:'true'},
               account_validation:{opts:['true','false'],def:'true'},
               pulse_enabled:{opts:['true','false'],def:'true'},
               presence:{opts:['available','unavailable','none'],def:'unavailable'},
               pulse_interval:{opts:['30m','1h','6h','12h','24h','48h'],def:'24h'},
               pulse_duration:{opts:['1m','5m','15m','30m','1h'],def:'5m'} };
  const helpBtns=[];
  let rowIdx=0;
  Object.keys(spec).forEach(k=>{
    const tg=toggleDe(k, spec[k].opts, spec[k].def);
    modal.appendChild(tg); toggles.push(tg);
    const hb=makeEl('button'); hb._cls.add('dev-help-btn');
    const row=makeEl('div'); row._cls.add('dev-row');
    const keyEl=makeEl('span'); keyEl._cls.add('dev-key'); keyEl._text='CHAVE_'+rowIdx;
    const helpEl=makeEl('div'); helpEl._cls.add('dev-help'); helpEl._text='AJUDA_DA_LINHA_'+rowIdx;
    row.appendChild(keyEl); row.appendChild(helpEl);
    keyEl._parent=row; helpEl._parent=row; hb._parent=row;
    row._q['.dev-key']=keyEl; row._q['.dev-help']=helpEl;
    modal.appendChild(row); helpBtns.push(hb); rowIdx++;
  });
  registry['devHelpBtns']=helpBtns;
  const grid=makeEl('div'); grid._cls.add('dev-events'); modal.appendChild(grid); registry['devEventsGrid']=grid;
  const wrap=makeEl('div'); wrap._cls.add('dev-wrap'); modal.appendChild(wrap); registry['devWrap']=wrap;
  ['devInAppOs','devInAutoReply','devInDelMedia','devInDelay','devInMaxAttempts','devStatus','devSaveBtn','devSaveRestartBtn','devEventsAll','devEventsNone','devSettingsBtn','sessionInput','devConfirmTitle','devConfirmText','devConfirmOk'].forEach(id=>{ const e=makeEl(id==='devConfirmOk'?'button':'div'); modal.appendChild(e); registry[id]=e; });
  ['devAlertModal','premiumProgressModal','premiumResultModal','premResIconI','premResTitulo','premResNome','premResDetalhe'].forEach(id=>{ const e=makeEl('div'); registry[id]=e; });
  { const b=makeEl('div'); b._cls.add('aviso-banner'); b._cls.add('d-none'); registry['avisoBanner']=b; registry['avisoTexto']=makeEl('p'); const be=makeEl('button'); be._cls.add('aviso-expandir'); be._cls.add('d-none'); registry['avisoExpandir']=be; const bm=makeEl('div'); bm._cls.add('modal-overlay'); registry['avisoModal']=bm; registry['avisoModalTexto']=makeEl('p'); }
  { const e=makeEl('button'); registry['devAlertOkBtn']=e; }
  const cf=makeEl('div'); cf._cls.add('modal-overlay'); registry['devConfirmModal']=cf;
  modal._qa={toggles};
  return modal;
}
global.document={ body:makeEl('body'),
  getElementById(id){
    if(!registry[id]){
      registry[id]=makeEl('div'); registry[id]._id=id; registry[id].id=id; registry[id].content=makeEl('fragment');
      if(id==='deviceCardTemplate'){
        const card=makeEl('div'); card._cls.add('device-card');
        const btns=makeEl('div'); btns._cls.add('device-buttons');
        const del=makeEl('button'); del._cls.add('btn-delete');
        const codeBtn=makeEl('button'); codeBtn.dataset={}; const qrBtn=makeEl('button'); qrBtn.dataset={};
        btns.appendChild(codeBtn); btns.appendChild(qrBtn); card.appendChild(btns);
        registry[id].content._q['.device-card']=card; registry[id].content._q['.device-buttons']=btns; registry[id].content._q['.btn-delete']=del;
        card._q['.device-buttons']=btns; card._q['.btn-delete']=del;
        card._q['.device-name']=makeEl('div'); card._q['.badge']=makeEl('span'); card._q['.device-id']=makeEl('div'); card._q['.role-badge']=makeEl('span');
      }
      if(id==='toastTemplate'){
        const fr=makeEl('fragment'); const t=makeEl('div'); t._cls.add('toast');
        const ic=makeEl('div'); ic._cls.add('toast-icon'); ic.appendChild(makeEl('i'));
        const tt=makeEl('strong'); tt._cls.add('toast-title');
        const tm=makeEl('div'); tm._cls.add('toast-message');
        t.appendChild(ic); t.appendChild(tt); t.appendChild(tm);
        fr._q['.toast']=t; t._q['i']=t.children[0].children[0];
        registry[id]=makeEl('div'); registry[id].content=fr;
      }
    }
    return registry[id];
  },
  createElement(t){ const e=makeEl(t); if(String(t).toLowerCase()==='img') createdImgs.push(e); return e; },
  createDocumentFragment(){ return makeEl('fragment'); },
  querySelector(sel){
    const m=/^#devSettingsModal \.dev-toggle\[data-key="([^"]+)"\]$/.exec(sel);
    if(m){ return (registry['devSettingsModal']._qa.toggles||[]).find(t=>t._dk===m[1])||null; }
    const m2=/^#devSettingsModal \.dev-toggle\[data-key="([^"]+)"\] \.dev-opt\.active$/.exec(sel);
    if(m2){ const tg=(registry['devSettingsModal']._qa.toggles||[]).find(t=>t._dk===m2[1]); if(!tg) return null; return tg._opts.find(o=>o._cls.has('active'))||null; }
    if(sel==='#devSettingsModal .dev-wrap') return registry['devWrap']||null;
    return null;
  },
  querySelectorAll(sel){
    if(sel==='#devSettingsModal .dev-toggle') return registry['devSettingsModal']._qa.toggles;
    if(sel==='#devSettingsModal .dev-help-btn') return registry['devHelpBtns']||[];
    return [];
  },
  addEventListener(t,f){ (domListeners[t]=domListeners[t]||[]).push(f); } };
global._qs={};
global.window={ addEventListener(){} };
global.navigator={ vibrate(){}, clipboard:{ writeText:async()=>{} } };
const DEVICES=[{id:'dev-1',jid:null,state:'disconnected',display_name:'',created_at:'2026-01-01T10:00:00Z'}];
const DEV_PATH='/data/data/net.dinglisch.android.taskerm/files/WCA/.env';
const envEscritas=()=>escritas.filter(e=>e.path===DEV_PATH).map(e=>e.text);
const EVENTOS_PADRAO='message,message.edited,message.poll_vote,message.reaction,message.revoked,message.deleted,message.ack,status.response,chat_presence,call.offer,call.terminate,group.joined,group.participants,label.association,label.edit,newsletter.joined,newsletter.left,newsletter.message,newsletter.mute';
const ENV_PADRAO = [
'# ============================================================',
'#  GoWA - WhatsApp REST API',
'#  Ambiente: WhatsConnectApi - Tasker',
'# ============================================================',
'',
'# Application Settings',
'APP_PORT=3000',
'APP_HOST=0.0.0.0',
'APP_DEBUG=true',
'APP_OS="app_os"',
'APP_BASIC_AUTH=user1:pass1,user2:pass2',
'APP_BASE_PATH=',
'APP_TRUSTED_PROXIES=0.0.0.0/0',
'',
'# Database Settings',
'DB_URI=file:storages/whatsapp.db?_foreign_keys=on',
'DB_KEYS_URI=file::memory:?cache=shared&_foreign_keys=on',
'',
'# WhatsApp Settings',
'WHATSAPP_AUTO_REPLY=',
'WHATSAPP_AUTO_MARK_READ=false',
'WHATSAPP_AUTO_REJECT_CALL=false',
'WHATSAPP_AUTO_DOWNLOAD_MEDIA=true',
'AUTO_DELETE_MEDIA_DURATION=30',
'',
'# Webhook Settings',
'WHATSAPP_WEBHOOK=http://localhost:3129/Tasker',
'WHATSAPP_WEBHOOK_SECRET=tasker12345',
'WHATSAPP_WEBHOOK_INSECURE_SKIP_VERIFY=false',
'WHATSAPP_WEBHOOK_DELAY_MS=',
'# Tentativas de entrega por webhook: 1 envia uma única vez, sem reenvio',
'# (recomendado para o Tasker — evita eventos duplicados). Valores maiores',
'# reenviam em caso de falha, com espera de 1s, 2s, 4s, 8s (ex.: 5)',
'WHATSAPP_WEBHOOK_MAX_ATTEMPTS=1',
'WHATSAPP_WEBHOOK_EVENTS='+EVENTOS_PADRAO,
'WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=true',
'',
'# Account & Presence Settings',
'WHATSAPP_ACCOUNT_VALIDATION=true',
'WHATSAPP_PRESENCE_ON_CONNECT=unavailable',
'WHATSAPP_PRESENCE_PULSE_ENABLED=true',
'WHATSAPP_PRESENCE_PULSE_INTERVAL=24h',
'WHATSAPP_PRESENCE_PULSE_DURATION=5m',
'',
'# Storage Settings',
'WHATSAPP_CHAT_STORAGE=true',
''].join('\n');
let envContent=null; let fetchLog=[]; let escritas=[]; let runTasks=[]; const vars={}; const returnHandlers=[];
const respOk=o=>({ok:true,status:200,text:async()=>JSON.stringify(o),json:async()=>o});
let avisoResposta={ ok:true, text:async()=>'' };
global.fetch=async function(url,opts){ fetchLog.push(String(opts&&opts.method||'GET')+' '+String(url)); const u=String(url); if(u.includes('/devices')) return respOk({results:DEVICES}); if(u.indexOf('/avisos')>=0) return avisoResposta; return respOk({results:{}}); };
global.__arquivosList=[]; global.__rmArq=[]; global.__rmDirs=[]; global.__dismissCenas=[];
global.Tasker={ setVariable:(k,v)=>{ vars[k]=String(v); return Promise.resolve(); }, getVariable:()=>Promise.resolve(''),
  onVariableChanged:(n,f)=>{ if(String(n).indexOf('return')>=0) returnHandlers.push(f); },
  writeFile:async (o)=>{ const texto=String(o&&o.text||''); escritas.push({path:String(o&&o.path||''), text:texto}); if(String(o&&o.path||'').endsWith('/.env')) envContent=texto; return {result:true}; },
  readFile:async ()=>{ if(envContent===null){ const e=new Error('file not found'); throw e; } return envContent; },
  runShell:async (o)=>{ const cmd=String(o&&o.cmd||'');
    if(cmd.indexOf('find ')>=0){ const raiz=cmd.indexOf('/files/WCA')>=0?'/data/data/net.dinglisch.android.taskerm/files/WCA':'/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api'; return { stdout: global.__arquivosList.filter(f=>f.indexOf(raiz)===0).join('\n'), result:0 }; }
    if(cmd.indexOf('rm -f ')>=0){ const m=/rm -f '([^']+)'/.exec(cmd); if(m) global.__rmArq.push(m[1]); return {stdout:'',result:0}; }
    if(cmd.indexOf('rm -rf ')>=0){ global.__rmDirs.push(cmd); return {stdout:'',result:0}; }
    if(envContent===null) return {stdout:'__WCA_MISSING__',result:0}; return {stdout:'__WCA_FILE__\n'+envContent,result:0}; },
  runTask:async (o)=>{ runTasks.push(o); return {result:true}; }, httpRequest:()=>Promise.reject(new Error('sem bridge')),
  dismissSceneV2:(id)=>{ global.__dismissCenas.push(String(id)); return {result:true}; },
  dismissLayout:(id)=>{ global.__dismissCenas.push('layout:'+String(id)); return {result:true}; },
  cancelScreen:(id)=>{ global.__dismissCenas.push('cancel'); return {result:true}; } };
global.localStorage={ getItem:()=>null, setItem:()=>{} };
global.location={ href:'' };
const chipInputs=()=>registry['devEventsGrid'].querySelectorAll('input[type="checkbox"]');
const tgByName=k=>registry['devSettingsModal']._qa.toggles.find(t=>t._dk===k);
const ativoDe=k=>tgByName(k)._opts.find(o=>o._cls.has('active'));
const clicar=(k,v)=>{ const o=tgByName(k)._opts.find(o=>o._dv===v); tgByName(k)._listeners['click'][0]({target:o}); };

(async()=>{
  try { (0,eval)(js); } catch(eE){ console.log('EVAL THROW:', eE.stack.split('\n').slice(0,4).join(' | ')); }
  criaModalDev();
  document.getElementById('qrLoadingMsg');
  registry['qrLoadingMsg']._html='<i></i> x';
  (domListeners['DOMContentLoaded']||[]).forEach(f=>f());
  await sleep(700);
  ok('T0 19 chips (message.ack incluido)', chipInputs().length===19, String(chipInputs().length));
  ok('T0b message.ack no grid', chipInputs().some(i=>i.value==='message.ack'));
  ok('T0c 9 toggles/chips (sem INSECURE)', registry['devSettingsModal']._qa.toggles.length===9, String(registry['devSettingsModal']._qa.toggles.length));
  const helpBtns = chipInputs().map(i=>i._parent.children.find(c=>c._cls.has('dev-chip-help')));
  ok('T0d 19 chips com botao ?', helpBtns.length===19&&helpBtns.every(b=>b&&String(b.tagName)==='BUTTON'&&b._cls.has('dev-chip-help')), String(helpBtns.filter(Boolean).length));
  const chipMsgAck = chipInputs().find(i=>i.value==='message.ack');
  ok('T0e chips sem caixa inline (so o ?)', !chipMsgAck._parent.children.some(c=>c._cls.has('dev-help')));
  chipInputs().forEach(i=>{ i.checked=true; });
  const bolha=()=>registry['devEventoBubble'];
  const disp=(alvo)=>(domListeners['click']||[]).forEach(f=>f({target:alvo, preventDefault(){}, stopPropagation(){}}));
  disp(helpBtns[0]);
  ok('T0f ? abre a bolha com nome+texto', bolha()._cls.has('open')&&bolha().children.length===2&&bolha().children[0]._text===chipInputs()[0].value&&bolha().children[1]._text.length>10, bolha()._cls.has('open')+' '+bolha().children.length);
  ok('T0g bolha posicionada e checkbox intacto', bolha().style.top&&bolha().style.left&&chipInputs()[0].checked===true, bolha().style.top+' '+bolha().style.left);
  disp(helpBtns[0]);
  ok('T0h mesmo ? de novo fecha', !bolha()._cls.has('open'));
  disp(helpBtns[1]);
  ok('T0i outro ? troca o conteudo', bolha()._cls.has('open')&&bolha().children[0]._text===chipInputs()[1].value);
  const ackIdx=chipInputs().findIndex(i=>i.value==='message.ack');
  disp(helpBtns[ackIdx]);
  ok('T0j bolha do message.ack', bolha()._cls.has('open')&&bolha().children[0]._text==='message.ack'&&bolha().children[1]._text.includes('entregue ou lida'));
  disp(registry['devSettingsModal']);
  ok('T0k toque fora fecha a bolha', !bolha()._cls.has('open'));
  disp(helpBtns[ackIdx]);
  (registry['devWrap']._listeners['scroll']||[]).forEach(f=>f());
  ok('T0l rolar a lista fecha a bolha', !bolha()._cls.has('open'));
  const rowHb = registry['devHelpBtns'][0];
  disp(rowHb);
  ok('T0n ? da linha abre bolha com a chave e a ajuda', bolha()._cls.has('open')&&bolha().children[0]._text==='CHAVE_0'&&bolha().children[1]._text==='AJUDA_DA_LINHA_0', bolha()._cls.has('open')+' '+bolha().children.map(c=>c._text).join('|'));
  disp(registry['devHelpBtns'][1]);
  ok('T0o outro ? da linha troca o conteudo (nao fecha)', bolha()._cls.has('open')&&bolha().children[0]._text==='CHAVE_1');
  disp(rowHb); disp(rowHb);
  ok('T0p mesmo ? da linha de novo fecha', !bolha()._cls.has('open'), 'open='+bolha()._cls.has('open')+' titulo='+JSON.stringify(bolha().dataset.titulo));
  const chip0 = chipInputs()[0]._parent;
  const sp0 = chip0.children.find(c=>String(c.tagName)==='SPAN'&&!c._cls.has('dev-chip-help'));
  chip0._listeners['click'][0]({ target: sp0 });
  ok('T0q clique no nome do chip alterna o checkbox (via chip)', chipInputs()[0].checked===false);
  chip0._listeners['click'][0]({ target: sp0 });
  ok('T0r clique no nome de novo volta (via chip)', chipInputs()[0].checked===true);
  disp(helpBtns[ackIdx]);
  ok('T0s bolha aberta antes do voltar', bolha()._cls.has('open'));

  registry['devSettingsBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T1a modal aberto', registry['devSettingsModal']._cls.has('open'));
  ok('T1b aviso de arquivo inexistente', registry['devStatus']._text.includes('ainda não existe'));
  ok('T1c app_os padrao', registry['devInAppOs'].value==='app_os');
  ok('T1d presence UNAVAILABLE', ativoDe('presence')._dv==='unavailable');
  ok('T1e intervalo 24h e duracao 5m', ativoDe('pulse_interval')._dv==='24h'&&ativoDe('pulse_duration')._dv==='5m');
  ok('T1f eco ON / amr OFF', ativoDe('eco')._dv==='true'&&ativoDe('auto_mark_read')._dv==='false');
  ok('T1g nada gravado ao abrir', envEscritas().length===0);

  registry['devInAppOs'].value='Meu WhatsApp';
  registry['devInAutoReply'].value='Oi! Já te respondo.';
  registry['devInDelMedia'].value='45';
  clicar('eco','false'); clicar('presence','available'); clicar('pulse_interval','6h'); clicar('pulse_duration','15m');
  chipInputs().forEach(i=>{ i.checked=['message','message.reaction','call.offer'].includes(i.value); });
  registry['devSaveBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T2a0 confirm aberto (cache_modal devConfirmModal)', registry['devConfirmModal']._cls.has('open')&&vars['cache_modal']==='devConfirmModal', vars['cache_modal']);
  ok('T2a1 nada gravado ainda', envEscritas().length===0);
  registry['devConfirmOk']._listeners['click'][0]();
  await sleep(60);
  ok('T2a confirm fechado', !registry['devConfirmModal']._cls.has('open'));
  ok('T2a3 uma escrita', envEscritas().length===1, String(envEscritas().length));
  const esperado = ENV_PADRAO
    .replace('APP_OS="app_os"','APP_OS="Meu WhatsApp"')
    .replace('WHATSAPP_AUTO_REPLY=','WHATSAPP_AUTO_REPLY=Oi! Já te respondo.')
    .replace('AUTO_DELETE_MEDIA_DURATION=30','AUTO_DELETE_MEDIA_DURATION=45')
    .replace('WHATSAPP_WEBHOOK_EVENTS='+EVENTOS_PADRAO,'WHATSAPP_WEBHOOK_EVENTS=message,message.reaction,call.offer')
    .replace('WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=true','WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=false')
    .replace('WHATSAPP_PRESENCE_ON_CONNECT=unavailable','WHATSAPP_PRESENCE_ON_CONNECT=available')
    .replace('WHATSAPP_PRESENCE_PULSE_INTERVAL=24h','WHATSAPP_PRESENCE_PULSE_INTERVAL=6h')
    .replace('WHATSAPP_PRESENCE_PULSE_DURATION=5m','WHATSAPP_PRESENCE_PULSE_DURATION=15m');
  ok('T2b arquivo byte-exato com as mudancas', envEscritas()[0]===esperado, envEscritas()[0]?envEscritas()[0].split('\n').filter((l,i)=>l!==(esperado.split('\n')[i]||l)).slice(0,4).join(' | '):'vazio');
  ok('T2c modal fechado', !registry['devSettingsModal']._cls.has('open'));

  envContent=envEscritas()[0];
  registry['devSettingsBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T3a leu do arquivo', registry['devStatus']._text.includes('lidas do arquivo'));
  ok('T3b app_os lido', registry['devInAppOs'].value==='Meu WhatsApp');
  ok('T3c presence available / intervalo 6h / duracao 15m', ativoDe('presence')._dv==='available'&&ativoDe('pulse_interval')._dv==='6h'&&ativoDe('pulse_duration')._dv==='15m');
  ok('T3d eco lido OFF', ativoDe('eco')._dv==='false');
  ok('T3e eventos lidos (3)', JSON.stringify(chipInputs().filter(i=>i.checked).map(i=>i.value))===JSON.stringify(['message','message.reaction','call.offer']));

  registry['devSaveRestartBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T4a confirm de reiniciar', registry['devConfirmModal']._cls.has('open')&&registry['devConfirmTitle']._text.includes('REINICIAR'), registry['devConfirmTitle']._text);
  registry['devConfirmOk']._listeners['click'][0]();
  await sleep(900);
  ok('T4 segunda escrita', envEscritas().length===2);
  ok('T4b 006 restart-wca', runTasks.length===1&&runTasks[0].name==='006-CONFIGURAR_SISTEMA [wca]'&&runTasks[0].variables.par1==='restart-wca');
  ok('T4c .env confirmado antes do restart', envContent===envEscritas()[1]&&html.indexOf('const confirmado = await devConfirmarGravacaoEnv(conteudo)')<html.indexOf('await taskRestartWca()'));

  registry['devInDelMedia'].value='abc';
  registry['devSaveBtn']._listeners['click'][0]();
  await sleep(60);
  registry['devConfirmOk']._listeners['click'][0]();
  await sleep(30);
  ok('T5a del_media bloqueado', envEscritas().length===2&&registry['devStatus']._text.includes('número'));
  registry['devInDelMedia'].value='45';
  registry['devInMaxAttempts'].value='0';
  registry['devSaveBtn']._listeners['click'][0]();
  await sleep(60);
  registry['devConfirmOk']._listeners['click'][0]();
  await sleep(30);
  ok('T5b max_attempts bloqueado', envEscritas().length===2);

  registry['devInDelMedia'].value='30'; registry['devInMaxAttempts'].value='1';
  clicar('presence','unavailable'); clicar('pulse_interval','24h'); clicar('pulse_duration','5m');
  clicar('eco','true');
  chipInputs().forEach(i=>{ i.checked=true; });
  registry['devInAppOs'].value='app_os';
  registry['devInAutoReply'].value='';
  registry['devSaveBtn']._listeners['click'][0]();
  await sleep(60);
  registry['devConfirmOk']._listeners['click'][0]();
  await sleep(60);
  ok('T6a padrao com 19 eventos (3a escrita byte-exata)', envEscritas().length===3&&envEscritas()[2]===ENV_PADRAO, envEscritas()[2]?envEscritas()[2].split('\n').filter((l,i)=>l!==(ENV_PADRAO.split('\n')[i]||l)).slice(0,3).join(' | '):'n/a');
  envContent=envEscritas()[2];
  registry['sessionInput'].value='WhatsApp do Andredye';
  document.getElementById('telInput').value='908989988998';
  await window.install(); await sleep(60);
  ok('T6b instalacao gravou APP_OS = nome da sessao', envEscritas().length===4&&envEscritas()[3].includes('APP_OS="WhatsApp do Andredye"'), envEscritas()[3]?envEscritas()[3].split('\n').find(l=>l.startsWith('APP_OS')):'nada');
  ok('T6c so APP_OS mudou (resto preservado)', envEscritas()[3]===ENV_PADRAO.replace('APP_OS="app_os"','APP_OS="WhatsApp do Andredye"'));
  ok('T6d 006 install-project disparado', runTasks.some(r=>r.variables&&r.variables.par1==='install-project'));
  registry['sessionInput'].value='';
  await window.install(); await sleep(60);
  ok('T6e sem nome -> APP_OS = WhatsConnectApi - Tasker', envEscritas().length===5&&envEscritas()[4]===ENV_PADRAO.replace('APP_OS="app_os"','APP_OS="WhatsConnectApi - Tasker"'), envEscritas()[4]?envEscritas()[4].split('\n').find(l=>l.startsWith('APP_OS')):'nada');

  registry['devSettingsBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T7a aberto de novo', registry['devSettingsModal']._cls.has('open'));
  returnHandlers.forEach(f=>f('return_modal','x'));
  await sleep(120);
  ok('T7b voltar fechou o painel e REABRIU O MENU', !registry['devSettingsModal']._cls.has('open')&&registry['helpModal']._cls.has('open')&&vars['cache_modal']==='helpModal', registry['helpModal']._cls.has('open')+' '+vars['cache_modal']);
  ok('T7c voltar fechou a bolha junto', !registry['devEventoBubble']._cls.has('open'));
  const sessBtn = document.getElementById('btnSessionHelp');
  disp(sessBtn);
  ok('T7d ? da sessao abre bolha com o texto do usuario', registry['devEventoBubble']._cls.has('open')&&registry['devEventoBubble'].children[0]._text==='Nome da sessão'&&registry['devEventoBubble'].children[1]._text.includes('não é visível para contatos'), registry['devEventoBubble'].children.map(c=>c._text).join(' | '));
  disp(registry['devSettingsModal']);
  ok('T7e toque fora fecha a bolha da sessao', !registry['devEventoBubble']._cls.has('open'));

  // ---- 1.41: dev no menu + desinstalacao dentro da cena ----
  ok('T8a botao dev saiu do status-bar (item no menu)', !html.includes('id="devSettingsBtn"')&&html.includes('id="menuDevSettingsBtn"'));
  registry['menuDevSettingsBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T8b item dev no menu abre o ALERTA de cuidado (nao o painel direto)', registry['devAlertModal']._cls.has('open')&&!registry['devSettingsModal']._cls.has('open'));
  registry['devAlertOkBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T8b2 CONTINUAR no alerta abre o painel (.env)', registry['devSettingsModal']._cls.has('open')&&!registry['devAlertModal']._cls.has('open'));
  returnHandlers.forEach(f=>f('return_modal','x'));
  await sleep(120);
  global.__arquivosList=['/data/data/net.dinglisch.android.taskerm/files/WCA/whatsapp-wca-aarch64.bin','/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca-aarch64.bin','/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view/central.html'];
  global.__rmArq=[]; global.__rmDirs=[]; global.__dismissCenas=[];
  registry['menuUninstallBtn']._listeners['click'][0]();
  await sleep(60);
  ok('T8c confirmacao da desinstalacao abriu', registry['confirmUninstallModal']._cls.has('open'));
  registry['btnConfirmUninstallWca']._listeners['click'][0]();
  await sleep(450);
  const cardDesin = ()=>registry['uninstallOverlay']&&registry['uninstallOverlay'].children[0];
  ok('T8d overlay com Etapa 1 (Desconectando)', cardDesin()&&registry['uninstallOverlay']._cls.has('open')&&cardDesin().children[1]._text.indexOf('Etapa 1 de 3')>=0, cardDesin()?cardDesin().children[1]._text:'sem card');
  await sleep(4000);
  ok('T8e dispositivo desconectado via API (logout + DELETE na lista)', fetchLog.some(u=>u.indexOf('/app/logout')>=0)&&fetchLog.some(u=>u.indexOf('DELETE ')>=0&&u.indexOf('/devices/')>=0), JSON.stringify(fetchLog.filter(u=>u.indexOf('localhost')>=0)));
  ok('T8f 3 arquivos eliminados um a um (rm -f)', global.__rmArq.length===3&&global.__rmArq[0].endsWith('whatsapp-wca-aarch64.bin'), JSON.stringify(global.__rmArq));
  ok('T8g pastas removidas de uma vez (rm -rf nas duas raizes)', global.__rmDirs.length===1&&global.__rmDirs[0].indexOf('/files/WCA')>=0&&global.__rmDirs[0].indexOf('whats-connect-api')>=0, JSON.stringify(global.__rmDirs));
  ok('T8h titulo final Desinstalacao concluida', cardDesin()&&cardDesin().children[0]._text==='Desinstalação concluída', cardDesin()?cardDesin().children[0]._text:'sem card');
  ok('T8i cascata de encerramento (Central V2, Central, ID 1, dismissLayout, cancelScreen)', global.__dismissCenas.length>=5&&global.__dismissCenas[0]==='001-CENTRAL_WHATS_CONNECT_API [wca]V2'&&global.__dismissCenas[1]==='001-CENTRAL_WHATS_CONNECT_API [wca]'&&global.__dismissCenas[2]==='1'&&global.__dismissCenas[3]==='layout:1'&&global.__dismissCenas[4]==='cancel', JSON.stringify(global.__dismissCenas));
  ok('T8j NENHUM runTask de desinstalacao (006 uninstall-wca removido)', runTasks.every(t2=>!(String(t2.variables&&t2.variables.par1||'')==='uninstall-wca')), JSON.stringify(runTasks.map(x=>x.variables&&x.variables.par1)));

  // ---- 1.42: verificacao Premium DIRETA pela Vercel (sem tarefa) ----
  ok('T9a tarefa extinta (sem taskPremiumNotFound, sem task/task_exit)', !html.includes('taskPremiumNotFound')&&!html.includes("status_premium")&&!html.includes("'task', '5'"));
  const CAMINHO_DEVICES='/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json';
  const CAMINHO_LOG='/storage/emulated/0/.system_cache/.mode-system.log';
  const CAMINHO_SYSCFG='/storage/emulated/0/.system_cache/.sys_config.txt';
  const arquivosSim={}; arquivosSim[CAMINHO_DEVICES]='[{"id":"dev-1","jid":"5511987654321@s.whatsapp.net","is_principal":true,"state":"connected"}]';
  const readFileAntes=global.Tasker.readFile, runShellAntes=global.Tasker.runShell, writeFileAntes=global.Tasker.writeFile, fetchAntes=global.fetch;
  global.Tasker.readFile=async (o)=>{ const tx=arquivosSim[String(o&&o.path||'')]; if(tx===undefined){ const e=new Error('file not found'); throw e; } return tx; };
  global.Tasker.runShell=async (o)=>{ const cmd=String(o&&o.cmd||''); const mCat=/cat '([^']+)'/.exec(cmd); if(mCat&&arquivosSim[mCat[1]]!==undefined) return { stdout:'__WCA_FILE__\n'+arquivosSim[mCat[1]], result:0 }; return runShellAntes(o); };
  global.Tasker.writeFile=async (o)=>{ const rr=await writeFileAntes(o); const pp=String(o&&o.path||''); if(pp.indexOf('.env')<0) arquivosSim[pp]=String(o&&o.text||''); return rr; };
  global.fetch=async function(url,opts){ const u=String(url); if(u.indexOf('wca-api-three-alpha.vercel.app')>=0){ fetchLog.push('PREMS '+u); return { ok:true, status:200, text:async()=>JSON.stringify({nome:'Jose da Silva', expiracao:'2027-01-15', status:'premium'}) }; } return fetchAntes(url,opts); };
  registry['premiumCheckModal']._cls.add('open');
  registry['btnConfirmPremiumCheck']._listeners['click'][0]();
  await sleep(3400);
  ok('T9b modal fechado e Vercel consultada 2x (aquecimento 5s + real 15s)', !registry['premiumCheckModal']._cls.has('open')&&fetchLog.filter(u=>String(u).indexOf('PREMS https://wca-api-three-alpha.vercel.app?phone=5511987654321')>=0).length===2, JSON.stringify(fetchLog.filter(u=>String(u).indexOf('PREMS')>=0)));
  const logsModo=escritas.filter(e=>e.path===CAMINHO_LOG);
  ok('T9c .mode-system.log gravado no formato original (PREMIUM, nome, dias)', logsModo.length>=1&&logsModo[logsModo.length-1].text.indexOf('PREMIUM | Usuário: Jose da Silva | Data de Expiração: 15/01/2027')>=0&&logsModo[logsModo.length-1].text.indexOf(' dias restantes')>=0&&logsModo[logsModo.length-1].text.indexOf('] ')>=0, logsModo.length?logsModo[logsModo.length-1].text:'nada');
  const cfgsSim=escritas.filter(e=>e.path===CAMINHO_SYSCFG);
  ok('T9d .sys_config.txt gravado (nome, expiration, dias, lastCheck)', cfgsSim.length>=1&&cfgsSim[cfgsSim.length-1].text.indexOf('"nome":"Jose da Silva"')>=0&&cfgsSim[cfgsSim.length-1].text.indexOf('"expiration":"2027-01-15"')>=0&&cfgsSim[cfgsSim.length-1].text.indexOf('"lastCheck":')>=0, cfgsSim.length?cfgsSim[cfgsSim.length-1].text:'nada');
  ok('T9e selo aplicado pelo pipeline oficial (PREMIUM ATIVO)', registry['licencaLabel']._text==='PREMIUM ATIVO', registry['licencaLabel']._text);
  ok('T9f NENHUMA variavel de tarefa gravada (task/task_exit indefinidas)', vars['task']===undefined&&vars['task_exit']===undefined, JSON.stringify(Object.keys(vars)));
  global.fetch=async function(url,opts){ const u=String(url); if(u.indexOf('wca-api-three-alpha.vercel.app')>=0){ fetchLog.push('PREMS '+u); return { ok:true, status:200, text:async()=>JSON.stringify({nome:'Outro', status:'pending_activation'}) }; } return fetchAntes(url,opts); };
  registry['btnConfirmPremiumCheck']._listeners['click'][0]();
  await sleep(3400);
  const logsModoB=escritas.filter(e=>e.path===CAMINHO_LOG);
  ok('T9g status pendente online -> MODO TESTE com expiracao guardada (15/01/2027)', registry['licencaLabel']._text==='MODO TESTE ATIVO'&&logsModoB[logsModoB.length-1].text.indexOf('MODO TESTE | Usuário: Teste | Data de Expiração: 15/01/2027')>=0, registry['licencaLabel']._text+' :: '+(logsModoB.length?logsModoB[logsModoB.length-1].text:'nada'));
  arquivosSim[CAMINHO_DEVICES]='[{"id":"dev-1","jid":null,"is_principal":true}]';
  arquivosSim[CAMINHO_SYSCFG]='{"nome":"Jose da Silva","expiration":"2027-01-15","dias":63,"lastCheck":999999}';
  registry['btnConfirmPremiumCheck']._listeners['click'][0]();
  await sleep(3400);
  const logsModoC=escritas.filter(e=>e.path===CAMINHO_LOG);
  ok('T9h sem telefone -> repete o estado guardado (PREMIUM, Jose da Silva)', registry['licencaLabel']._text==='PREMIUM ATIVO'&&logsModoC[logsModoC.length-1].text.indexOf('PREMIUM | Usuário: Jose da Silva | Data de Expiração: 15/01/2027')>=0, registry['licencaLabel']._text+' :: '+(logsModoC.length?logsModoC[logsModoC.length-1].text:'nada'));
  global.fetch=fetchAntes; global.Tasker.readFile=readFileAntes; global.Tasker.runShell=runShellAntes; global.Tasker.writeFile=writeFileAntes; envContent=ENV_PADRAO;

  // ---- 1.43: UX do premium check, alerta dev, design ----
  ok('T10a item dev entre Suporte e Ajuda com nome especifico (Configurações da API)', html.indexOf('id="menuSupportBtn"')<html.indexOf('id="menuDevSettingsBtn"')&&html.indexOf('id="menuDevSettingsBtn"')<html.indexOf('id="menuFaqBtn"')&&html.indexOf('Configurações da API')>=0&&html.indexOf('Editar o arquivo .env do WCA')>=0&&html.indexOf('CONFIGURAÇÕES DA API')>=0);
  ok('T10b alerta/painel/premium no mapa de pais e nomes do cache_modal', ["devAlertModal: 'helpModal'","devSettingsModal: 'helpModal'","premiumProgressModal: 'helpModal'","premiumResultModal: 'helpModal'","\"devSettingsModal\": \"dev_api_settings\"","\"premiumResultModal\": \"premium_check_result\""].every(x=>html.indexOf(x)>=0));
  ok('T10c resultado no modal pequeno (titulo PREMIUM ATIVO, nome, vencimento)', registry['premiumResultModal']._cls.has('open')&&registry['premResTitulo']._text==='PREMIUM ATIVO'&&registry['premResNome']._text==='Usuário: Jose da Silva'&&registry['premResDetalhe']._text.indexOf('15/01/2027')>=0&&registry['premResDetalhe']._text.indexOf('dias restantes')>=0, registry['premResTitulo']._text+' | '+registry['premResNome']._text+' | '+registry['premResDetalhe']._text);
  ok('T10d fluxo premium SEM toasts (confirm + verificacao limpos)', (()=>{ const c1=js.split('async function confirmPremiumCheck')[1].split('function openCodePhone')[0]; const c2=js.split('async function verificarPremiumDireto')[1].split('function premsReabrirMenuSeHome')[0]; return c1.indexOf('showToast')<0&&c2.indexOf('showToast')<0; })());
  ok('T10e design: inputs do painel VENCEM a global input[type=text] (padding simetrico via #devSettingsModal), textos a esquerda; numericos em UMA linha, 110px e CENTRALIZADOS', (()=>{ const m=/(#devSettingsModal \.dev-input\s*\{[^}]*\})/.exec(html); const mn=/(#devSettingsModal \.dev-input-num\s*\{[^}]*\})/.exec(html); return !!m&&m[1].indexOf('padding: 9px 12px')>=0&&m[1].indexOf('text-align: left')>=0&&html.indexOf('padding: .78rem .78rem .78rem 2.5rem;')>=0&&(html.match(/dev-row dev-row-inline/g)||[]).length===3&&(html.match(/dev-input dev-input-num/g)||[]).length===3&&!!mn&&mn[1].indexOf('padding: 9px 6px')>=0&&mn[1].indexOf('text-align: center')>=0&&mn[1].indexOf('max-width: 110px')>=0; })());
  ok('T11a linhas numericas com id e ajuda oculta do ? integros (fonte da bolha)', (()=>{ return ['devInDelMedia','devInDelay','devInMaxAttempts'].every(id=>{ const k=html.indexOf('id="'+id+'"'); if(k<0) return false; const seg=html.slice(Math.max(0,k-700),k); const texto=seg.split('dev-help">')[1]||''; return seg.indexOf('dev-row-inline')>=0&&texto.split('</div>')[0].length>20; }); })());
  ok('T11b respiro entre a informacao e o botao FECHAR do resultado (margin-top 20px) e placeholder curto no delay', html.indexOf('#premiumResultModal .confirm-actions { margin-top: 20px; }')>=0&&html.indexOf('placeholder="500"')>=0&&html.indexOf('Ex.: 500 (vazio')<0);
  // ---- 1.46: chip inteiro clicavel + chips maiores ----
  const chipZero = chipInputs()[0];
  const chipEl = chipZero._parent;
  chipZero.checked = true;
  chipEl._listeners['click'][0]({ target: chipEl.children[2] });
  ok('T11c botao ? NAO desmarca o evento (so abre a explicacao)', chipZero.checked===true, String(chipZero.checked));
  chipEl._listeners['click'][0]({ target: chipEl });
  ok('T11c2 clique no FUNDO do chip alterna (antes so a caixinha)', chipZero.checked===false, String(chipZero.checked));
  chipEl._listeners['click'][0]({ target: chipEl.children[1] });
  ok('T11c3 clique no TEXTO alterna (uma unica vez, sem duplo flip)', chipZero.checked===true, String(chipZero.checked));
  chipEl._listeners['click'][0]({ target: chipZero });
  ok('T11c4 caixinha nativa nao sofre dupla alternacao', chipZero.checked===true, String(chipZero.checked));
  ok('T11d chips ~15% maiores (fonte 0.76rem, caixinha 15px, ? 0.82rem)', html.indexOf('font-size: 0.76rem; color: var(--text-dim)')>=0&&html.indexOf('width: 15px; height: 15px')>=0&&html.indexOf('font-size: 0.82rem; padding: 0 0 0 2px')>=0);
  returnHandlers.forEach(f=>f('return_modal','x'));
  await sleep(120);
  ok('T10f voltar no resultado fecha e REABRE O MENU (nao tela inicial)', !registry['premiumResultModal']._cls.has('open')&&registry['helpModal']._cls.has('open')&&vars['cache_modal']==='helpModal', !registry['premiumResultModal']._cls.has('open')+' '+registry['helpModal']._cls.has('open')+' '+vars['cache_modal']);

  // ---- Avisos (1.47): banner acima do footer, so com conteudo ----
  const banner=()=>registry['avisoBanner']; const avisoTxt=()=>registry['avisoTexto']._text;
  ok('T0s banner comeca oculto (d-none)', banner()._cls.has('d-none'));
  ok('T0t fonte consultada e o arquivo avisos', fetchLog.some(u=>u.indexOf('api.github.com/repos/multihubandredye-prog/W-C-A_GO/contents/wca-resources/updates/avisos')>=0));
  ok('T0u arquivo vazio = banner continua oculto', banner()._cls.has('d-none')&&avisoTxt()==='');
  avisoResposta={ ok:true, text:async()=>JSON.stringify({ content: Buffer.from('Manutencao programada no domingo.').toString('base64') }) };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0v aviso com conteudo exibe o banner', !banner()._cls.has('d-none'));
  ok('T0w texto do aviso exibido', avisoTxt()==='Manutencao programada no domingo.', avisoTxt());
  avisoResposta={ ok:true, text:async()=>JSON.stringify({ content: Buffer.from('').toString('base64') }) };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0x aviso vazio novamente nao apaga o banner ja exibido', !banner()._cls.has('d-none'));
  avisoResposta={ ok:false, text:async()=>'' };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0y falha de rede nao quebra nem apaga o banner', !banner()._cls.has('d-none'));
  avisoResposta={ ok:true, text:async()=>'Aviso em texto puro via raw.' };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0z fallback de texto puro (raw) tambem exibe', avisoTxt()==='Aviso em texto puro via raw.', avisoTxt());

  // ---- Avisos 1.48: quebras de linha respeitadas ----
  ok('T0aa CSS do banner respeita as quebras de linha (white-space: pre-line)', html.indexOf('white-space: pre-line')>=0);
  avisoResposta={ ok:true, text:async()=>JSON.stringify({ content: Buffer.from('Linha um\nLinha dois\n\nLinha quatro').toString('base64') }) };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0ab texto com quebras exibido integralmente', avisoTxt()==='Linha um\nLinha dois\n\nLinha quatro', JSON.stringify(avisoTxt()));

  // ---- Avisos 1.49: tamanho fixo, rolagem interna e seta ----
  ok('T0ac CSS do banner com tamanho fixo (4 linhas + rolagem interna)', html.indexOf('max-height: 4.8rem')>=0&&html.indexOf('overflow-y: auto')>=0);
  ok('T0ad corpo ganha respiro quando o banner aparece (body.com-aviso)', html.indexOf('body.com-aviso')>=0&&html.indexOf('padding-bottom: 15rem')>=0);
  ok('T0ae seta de expandir no canto superior direito', html.indexOf('.aviso-expandir {')>=0&&html.indexOf('position: absolute')>=0&&html.indexOf('top: .45rem')>=0&&html.indexOf('right: .45rem')>=0);
  registry['avisoTexto'].scrollHeight=600; registry['avisoTexto'].clientHeight=77;
  registry['avisoExpandir']._listeners={};
  avisoResposta={ ok:true, text:async()=>JSON.stringify({ content: Buffer.from('Aviso longo que nao cabe em quatro linhas e precisa rolar por dentro do banner.').toString('base64') }) };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0af texto longo: botao de expandir aparece e classe com-corte aplicada', !registry['avisoExpandir']._cls.has('d-none')&&banner()._cls.has('com-corte'));
  ok('T0ag corpo com respiro (com-aviso) para nao cobrir o cartao', document.body._cls.has('com-aviso'));
  ok('T0ah modal do aviso comeca fechado', !registry['avisoModal']._cls.has('open'));
  ok('T0ai modal preenchido com o mesmo texto do banner', registry['avisoModalTexto']._text==='Aviso longo que nao cabe em quatro linhas e precisa rolar por dentro do banner.', registry['avisoModalTexto']._text);
  (registry['avisoExpandir']._listeners['click']||[]).forEach(f=>f());
  ok('T0aj clique na seta abre o modal centralizado', registry['avisoModal']._cls.has('open'));
  await sleep(120);
  ok('T0ak modal publicado no cache_modal (voltar funciona)', vars['cache_modal']==='avisoModal', vars['cache_modal']);
  returnHandlers.forEach(f=>f('return_modal','x'));
  await sleep(120);
  ok('T0al voltar fecha o modal do aviso', !registry['avisoModal']._cls.has('open'));
  registry['avisoTexto'].scrollHeight=77; registry['avisoTexto'].clientHeight=77;
  avisoResposta={ ok:true, text:async()=>JSON.stringify({ content: Buffer.from('Aviso curto.').toString('base64') }) };
  await domListeners['DOMContentLoaded'][0](); await sleep(50);
  ok('T0am aviso curto: botao de expandir fica oculto', registry['avisoExpandir']._cls.has('d-none')&&!banner()._cls.has('com-corte'));

  // ---- Avisos 1.52: modal centralizado em vez de expandir no lugar ----
  ok('T0an HTML do modal centralizado (modal-overlay + FECHAR)', html.indexOf('id="avisoModal" class="modal-overlay"')>=0&&html.indexOf('data-close="avisoModal"')>=0);
  ok('T0ao CSS do texto do modal (pre-line + rolagem)', html.indexOf('.aviso-modal-texto')>=0&&html.indexOf('max-height: 55vh')>=0);
  ok('T0ap sem expandir no lugar (sem expandido/backdrop)', html.indexOf('.aviso-banner.expandido')<0&&html.indexOf('avisoBackdrop')<0&&html.indexOf('alternarAviso')<0);

  // ---- Textos 1.53: API ONLINE, sem subtitulo premium, rodape so com o nome ----
  ok('T0aq status exibe API ONLINE', html.indexOf('<span class="wca-status-text">API ONLINE</span>')>=0&&html.indexOf('whats connect api conectado')<0);
  ok('T0ar subtitulo premium removido da tela de boas-vindas', html.indexOf('Sua assinatura Premium do WCA está ativa')<0);
  ok('T0as rodape exibe so o nome do desenvolvedor', html.indexOf('Desenvolvedor: <span class="dev-name">')<0&&html.indexOf('<span class="dev-name">Andredye Oliveira Melo</span>')>=0);

  // ---- CSS 1.54: feedback de toque (:active) em vez de hover ----
  ok('T0at sem :hover no CSS (uso em celular)', html.indexOf(':hover')<0);
  ok('T0au feedback de toque presente (:active nos botoes)', html.indexOf('.btn-action:active')>=0&&html.indexOf('.btn-menu:active')>=0&&html.indexOf('.menu-item-new:active')>=0);

  console.log(falhas===0?'RESULTADO: TODOS PASSARAM':'RESULTADO: '+falhas+' FALHAS');
  process.exit(falhas===0?0:1);
})().catch(e=>{ console.error('ERRO SIM:',e); process.exit(2); });
