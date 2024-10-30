<?php
/*
 * MIT License
 *
 * Copyright (c) 2021-2024 machinateur
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

declare(strict_types=1);

namespace Machinateur\ChromeTabTransfer\FileTemplate;

use Machinateur\ChromeTabTransfer\Platform;

class ReopenScriptFileTemplate extends AbstractFileTemplate
{
    public function __construct(
        array                   $jsonArray,
        private readonly int    $port,
        private readonly string $socket,
    ) {
        parent::__construct($jsonArray);
    }

    public function getFileExtension(): string
    {
        return Platform::isWindows()
            ? 'cmd'
            : 'sh';
    }

    public function render(): string
    {
        $port   = $this->port;
        $socket = $this->socket;

        return Platform::isWindows()
            // Windows `cmd` script
            ? (
                '@echo off'
                . \PHP_EOL
                . \PHP_EOL
                . ':: Created using machinateur/android-chrome-tab-transfer (https://github.com/machinateur/android-chrome-tab-transfer).'
                . \PHP_EOL
                . \PHP_EOL
                . "adb -d forward tcp:{$port} localabstract:{$socket}"
                . \PHP_EOL
                . \PHP_EOL
                . \join(
                    \PHP_EOL, \array_map(static function (array $entry) use ($port): string {
                        $url   = $entry['url'];
                        $title = $entry['title'] ?: $url;
                        $url   = \rawurlencode($url);
                        // Batch files require all `%` to be escaped with another `%`.
                        $url = \str_replace('%', '%%', $url);

                        return \sprintf('# %s', $title)
                            . \PHP_EOL
                            // URLs must be double-quoted to work with curl, and double-quotes also escape other batch characters.
                            . \sprintf("curl -X PUT \"http://localhost:{$port}/json/new?%s\"", $url)
                            . \PHP_EOL;
                    }, $this->jsonArray)
                )
                . \PHP_EOL
                . \PHP_EOL
                . "adb -d forward --remove tcp:{$port}"
                . \PHP_EOL
                . \PHP_EOL
                . 'pause'
                . \PHP_EOL
                . \PHP_EOL
            )
            // Linux/Mac bash script
            : (
                '#!/bin/bash'
                . \PHP_EOL
                . \PHP_EOL
                . '# Created using machinateur/android-chrome-tab-transfer (https://github.com/machinateur/android-chrome-tab-transfer).'
                . \PHP_EOL
                . \PHP_EOL
                . "adb -d forward tcp:{$port} localabstract:{$socket}"
                . \PHP_EOL
                . \PHP_EOL
                . \join(
                    \PHP_EOL, \array_map(static function (array $entry) use ($port): string {
                        $url   = $entry['url'];
                        $title = $entry['title'] ?: $url;
                        $url   = \rawurlencode($url);

                        return \sprintf('# %s', $title)
                            . \PHP_EOL
                            . \sprintf("curl -X PUT 'http://localhost:{$port}/json/new?%s'", $url)
                            . \PHP_EOL;
                    }, $this->jsonArray)
                )
                . \PHP_EOL
                . \PHP_EOL
                . "adb -d forward --remove tcp:{$port}"
                . \PHP_EOL
            )
        ;
    }
}
