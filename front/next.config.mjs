import path from 'path';
import dotenv from 'dotenv';

// プロジェクトルートの一個上の階層にある.envファイルを読み込む
dotenv.config({ path: path.resolve(process.cwd(), '../.env') });

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
};

export default nextConfig;
