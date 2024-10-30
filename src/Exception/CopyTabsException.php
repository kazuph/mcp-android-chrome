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

class CopyTabsException extends \Exception
{
    public static function fromTabLoadingFailedException(TabLoadingFailedException $exception): self
    {
        // TODO: Check $exception contents to refine message directly in here, instead of in the command code.

        return new self('Failed to copy tabs from device!', previous: $exception);
    }

    public static function forEmptyResult(string $url): self
    {
        // TODO: Use $url in error message.

        return new self('Failed to copy tabs from device! Empty result.');
    }

    public static function fromFileTemplateDumpException(FileTemplateDumpException $exception): self
    {
        return new self('Failed to copy tabs from device! ' . $exception->getMessage(), previous: $exception);
    }
}
