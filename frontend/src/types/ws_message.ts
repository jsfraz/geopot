import type { Connection } from './connection';

export interface WSMessage {
  type: 'attacker' | 'server_info';
  data: Connection;
}