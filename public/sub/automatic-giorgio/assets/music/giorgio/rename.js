const fs = require('fs')


let dir =fs.readdirSync('./').filter(file => file.endsWith('.ogg'))
dir.forEach((file,i) => fs.renameSync(file, 'g'+i+'.ogg'))
