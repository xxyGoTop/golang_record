const go = new Go();

window.onload = function() {
    const env = {
        memoryBase: 0,
        tableBase: 0,
        memory: new WebAssembly.Memory({
          initial: 256
        }),
        table: new WebAssembly.Table({
          initial: 2,
          element: 'anyfunc'
        })
    };
    
    fetch("./main.wasm").then(response => 
        response.arrayBuffer()    
    ).then(bytes => 
        WebAssembly.instantiate(bytes, go.importObject)    
    ).then(result => {
        go.run(result.instance);
    }).catch(err => {
        console.log(err)
    })
    // WebAssembly.instantiateStreaming(fetch("./main.wasm"), go.importObject).then(result => {
    //     go.run(result.instance);
    // });
}