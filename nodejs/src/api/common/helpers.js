import fs from 'fs';

async function writeFile(filePath, data) {
    return await new Promise(async (resolve, reject) => {
        const dataString = JSON.stringify(data, null, 2);
        fs.writeFile(filePath, dataString, () => resolve());
    });
}

async function limitConcurrency() {

}

export {
    writeFile,
};
