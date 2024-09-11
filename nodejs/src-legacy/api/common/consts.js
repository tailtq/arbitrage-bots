import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const ROOT_DIR = path.resolve(path.dirname(__filename), '..');
const DATA_DIR = path.resolve(ROOT_DIR, 'data');
const LOGS_DIR = path.resolve(ROOT_DIR, 'logs');

export {
    ROOT_DIR,
    DATA_DIR,
    LOGS_DIR,
};
