import { ParamsBase } from './ParamsBase';

export interface ParamsStream extends ParamsBase {
    udid: string;
    player: string;
    ws?: string; // WebSocket URL for proxy-adb connection
}
