import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const ROOT_DIR = path.resolve(path.dirname(__filename), '..');
const LOGS_DIR = path.resolve(ROOT_DIR, 'logs');

export {
    ROOT_DIR,
    LOGS_DIR,
};
