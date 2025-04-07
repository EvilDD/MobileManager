import { BaseCanvasBasedPlayer } from './BaseCanvasBasedPlayer';
import VideoSettings from '../VideoSettings';
import Size from '../Size';
import { DisplayInfo } from '../DisplayInfo';
import H264Parser from 'h264-converter/dist/h264-parser';
import NALU from 'h264-converter/dist/util/NALU';
import ScreenInfo from '../ScreenInfo';
import Rect from '../Rect';

type ParametersSubSet = {
    codec: string;
    width: number;
    height: number;
};

function toHex(value: number) {
    return value.toString(16).padStart(2, '0').toUpperCase();
}

export class WebCodecsPlayer extends BaseCanvasBasedPlayer {
    public static readonly storageKeyPrefix = 'WebCodecsPlayer';
    public static readonly playerFullName = 'WebCodecs';
    public static readonly playerCodeName = 'webcodecs';

    public static readonly preferredVideoSettings: VideoSettings = new VideoSettings({
        lockedVideoOrientation: -1,
        bitrate: 5024288,
        maxFps: 24,
        iFrameInterval: 5,
        bounds: new Size(540, 960),
        sendFrameMeta: false,
    });

    public static isSupported(): boolean {
        if (typeof VideoDecoder !== 'function' || typeof VideoDecoder.isConfigSupported !== 'function') {
            return false;
        }

        // FIXME: verify support
        // const result = await VideoDecoder.isConfigSupported();
        return true;
    }

    private static parseSPS(data: Uint8Array): ParametersSubSet {
        const {
            profile_idc,
            constraint_set_flags,
            level_idc,
            pic_width_in_mbs_minus1,
            frame_crop_left_offset,
            frame_crop_right_offset,
            frame_mbs_only_flag,
            pic_height_in_map_units_minus1,
            frame_crop_top_offset,
            frame_crop_bottom_offset,
            sar,
        } = H264Parser.parseSPS(data);

        const sarScale = sar[0] / sar[1];
        const codec = `avc1.${[profile_idc, constraint_set_flags, level_idc].map(toHex).join('')}`;
        const width = Math.ceil(
            ((pic_width_in_mbs_minus1 + 1) * 16 - frame_crop_left_offset * 2 - frame_crop_right_offset * 2) * sarScale,
        );
        const height =
            (2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16 -
            (frame_mbs_only_flag ? 2 : 4) * (frame_crop_top_offset + frame_crop_bottom_offset);
        return { codec, width, height };
    }

    public readonly supportsScreenshot = true;
    private context: CanvasRenderingContext2D;
    private decoder: VideoDecoder;
    private buffer: ArrayBuffer | undefined;
    private hadIDR = false;
    private bufferedSPS = false;
    private bufferedPPS = false;

    constructor(udid: string, displayInfo?: DisplayInfo, name = WebCodecsPlayer.playerFullName) {
        super(udid, displayInfo, name, WebCodecsPlayer.storageKeyPrefix);
        const context = this.tag.getContext('2d');
        if (!context) {
            throw Error('Failed to get 2d context from canvas');
        }
        this.context = context;
        this.decoder = this.createDecoder();
    }

    private createDecoder(): VideoDecoder {
        return new VideoDecoder({
            output: (frame) => {
                this.onFrameDecoded(0, 0, frame);
            },
            error: (error: DOMException) => {
                console.error(error, `code: ${error.code}`);
                this.stop();
            },
        });
    }

    protected addToBuffer(data: Uint8Array): Uint8Array {
        let array: Uint8Array;
        if (this.buffer) {
            array = new Uint8Array(this.buffer.byteLength + data.byteLength);
            array.set(new Uint8Array(this.buffer));
            array.set(new Uint8Array(data), this.buffer.byteLength);
        } else {
            array = data;
        }
        this.buffer = array.buffer;
        return array;
    }

    protected scaleCanvas(width: number, height: number): void {
        console.log('原始视频尺寸:', { width, height });
        
        // 获取设备像素比
        const dpr = window.devicePixelRatio || 1;
        console.log('设备像素比:', dpr);
        
        // 计算实际显示尺寸
        const displayWidth = Math.round(width);
        const displayHeight = Math.round(height);
        
        // 计算Canvas的物理像素尺寸
        const canvasWidth = Math.round(displayWidth * dpr);
        const canvasHeight = Math.round(displayHeight * dpr);
        
        console.log('显示尺寸:', { displayWidth, displayHeight });
        console.log('Canvas物理像素尺寸:', { canvasWidth, canvasHeight });
        
        const screenInfo = new ScreenInfo(new Rect(0, 0, width, height), new Size(displayWidth, displayHeight), 0);
        this.emit('input-video-resize', screenInfo);
        this.setScreenInfo(screenInfo);

        // 初始化canvas，使用物理像素尺寸
        this.initCanvas(canvasWidth, canvasHeight);

        // 设置统一的CSS显示尺寸
        const commonStyle = {
            position: 'absolute',
            left: '0',
            top: '0',
            width: `${displayWidth}px`,
            height: `${displayHeight}px`,
            transformOrigin: '0 0',
            minWidth: '200px',
            minHeight: '200px'
        };

        // 应用样式到视频层
        Object.assign(this.tag.style, commonStyle);

        // 应用样式到触摸层
        Object.assign(this.touchableCanvas.style, commonStyle);
        
        // 设置Canvas的缩放以匹配设备像素比
        this.context.scale(dpr, dpr);
        
        // 如果父元素存在，也更新其尺寸
        if (this.parentElement) {
            this.parentElement.style.width = `${displayWidth}px`;
            this.parentElement.style.height = `${displayHeight}px`;
        }
    }

    protected decode(data: Uint8Array): void {
        if (!data || data.length < 4) {
            return;
        }
        const type = data[4] & 31;
        const isIDR = type === NALU.IDR;

        if (type === NALU.SPS) {
            const { codec, width, height } = WebCodecsPlayer.parseSPS(data.subarray(4));
            this.scaleCanvas(width, height);
            const config: VideoDecoderConfig = {
                codec,
                optimizeForLatency: true,
            } as VideoDecoderConfig;
            this.decoder.configure(config);
            this.bufferedSPS = true;
            this.addToBuffer(data);
            this.hadIDR = false;
            return;
        } else if (type === NALU.PPS) {
            this.bufferedPPS = true;
            this.addToBuffer(data);
            return;
        } else if (type === NALU.SEI) {
            // Workaround for lonely SEI from ws-qvh
            if (!this.bufferedSPS || !this.bufferedPPS) {
                return;
            }
        }
        const array = this.addToBuffer(data);
        this.hadIDR = this.hadIDR || isIDR;
        if (array && this.decoder.state === 'configured' && this.hadIDR) {
            this.buffer = undefined;
            this.bufferedPPS = false;
            this.bufferedSPS = false;
            this.decoder.decode(
                new EncodedVideoChunk({
                    type: 'key',
                    timestamp: 0,
                    data: array.buffer,
                }),
            );
            return;
        }
    }

    protected drawDecoded = (): void => {
        if (this.receivedFirstFrame) {
            const data = this.decodedFrames.shift();
            if (data) {
                const frame: VideoFrame = data.frame;
                this.context.drawImage(frame, 0, 0);
                frame.close();
            }
        }
        if (this.decodedFrames.length) {
            this.animationFrameId = requestAnimationFrame(this.drawDecoded);
        } else {
            this.animationFrameId = undefined;
        }
    };

    protected dropFrame(frame: VideoFrame): void {
        frame.close();
    }

    public getFitToScreenStatus(): boolean {
        // 总是返回true以确保视频适应屏幕
        return true;
    }

    public getPreferredVideoSetting(): VideoSettings {
        return WebCodecsPlayer.preferredVideoSettings;
    }

    public loadVideoSettings(): VideoSettings {
        return WebCodecsPlayer.loadVideoSettings(this.udid, this.displayInfo);
    }

    protected needScreenInfoBeforePlay(): boolean {
        return false;
    }

    public stop(): void {
        super.stop();
        if (this.decoder.state === 'configured') {
            this.decoder.close();
        }
    }
}
