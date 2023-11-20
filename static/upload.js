function uploadFile() {
    const fileInput = document.getElementById('fileInput');
    const resp = document.getElementById('response');
    const file = fileInput.files[0];

    // get filename with extension
    const filenameExt = file.name;
    console.log(filenameExt);

    if (!file) {
        alert("Please select a file first!");
        return;
    }
    updateProgressBar(0); // Reset progress bar

    const chunkSize = 1024 * 1024 * 5; // 5MB chunk size
    let currentChunk = 0;
    const totalChunks = Math.ceil(file.size / chunkSize);

    function uploadNextChunk() {
        const start = currentChunk * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end); // Getting a chunk

        // FormData to send the chunk
        const formData = new FormData();
        formData.append("file", chunk);
        formData.append("chunkIndex", currentChunk);
        formData.append("totalChunks", totalChunks);
        formData.append("filenameWithExtension", filenameExt);

        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/upload_chunk', true);

        // Update progress bar
        xhr.upload.onprogress = function(e) {
            if (e.lengthComputable) {
                const percentComplete = (e.loaded / e.total) * 100;
                updateProgressBar(percentComplete);
            }
        };

        xhr.onload = function() {
            if (xhr.status === 200) {
                console.log('Chunk ' + currentChunk + ' uploaded');
                resp.value += 'Chunk ' + currentChunk + ' uploaded' + '\n';
                currentChunk++;

                if (currentChunk < totalChunks) {
                    uploadNextChunk();
                } else {
                    console.log('Upload complete');
                    resp.value += 'Upload complete' + '\n';
                    updateProgressBar(100); // Complete
                }
            } else {
                console.error('Upload error: ' + xhr.statusText);
            }
        };

        xhr.send(formData);
    }

    uploadNextChunk();
}

function updateProgressBar(percent) {
    const progressBar = document.getElementById('progressBar');
    progressBar.style.width = percent + '%';
}
