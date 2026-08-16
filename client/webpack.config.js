const path = require('path');

module.exports = {
  // 为 @file-viewer 依赖添加 fallback
  resolve: {
    fallback: {
      "path": require.resolve("path-browserify"),
      "fs": false,
      "zlib": require.resolve("browserify-zlib"),
      "util": require.resolve("util")
    }
  },
  module: {
    rules: [
      {
        test: /\.wasm$/,
        type: "webassembly/async"
      }
    ]
  },
  experiments: {
    asyncWebAssembly: true
  }
};
