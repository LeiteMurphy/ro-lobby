// Servidor de produção do web (spec endurecimento, RN-11 e RN-12). O adapter-node entrega os
// arquivos estáticos (/_app, /robots.txt, imagens) antes do hook do SvelteKit; envolvendo o
// handler dele aqui, toda resposta sai com os cabeçalhos de segurança. Fora isso, faz o
// mesmo que o `node build`: lê PORT e HOST e encerra limpo no SIGTERM do Docker.
import http from 'node:http';
import { handler } from './build/handler.js';
import { headersFor, withHeaders } from './security-headers.js';

const port = Number(process.env.PORT ?? 3000);
const host = process.env.HOST ?? '0.0.0.0';
const https = (process.env.ORIGIN ?? '').startsWith('https://');

const server = http.createServer(withHeaders(handler, headersFor(https)));
server.listen(port, host, () => console.log(`web ouvindo em http://${host}:${port}`));

for (const signal of ['SIGTERM', 'SIGINT']) {
	process.on(signal, () => server.close(() => process.exit(0)));
}
