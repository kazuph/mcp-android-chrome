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

namespace Machinateur\ChromeTabTransfer\Exception;

class JsonFileLoadingException extends TabLoadingFailedException
{
    public static function whenFileNotFound(string $file): self
    {
        return new self("Input file `{$file}` not found.");
    }

    public static function whenFileNotAJsonFile(string $file): self
    {
        return new self("Input file `{$file}` is not a `.json` file.");
    }

    public static function whenFileNotReadable(string $file): self
    {
        return new self("Input file `{$file}` is not readable.");
    }

    public static function fromJsonException(\JsonException $exception): self
    {
        return new self("Failed decoding JSON content with error: {$exception->getMessage()}.", previous: $exception);
    }

    public static function forMalformedContent(): self
    {
        return new self('Detected malformed JSON content.');
    }
}
