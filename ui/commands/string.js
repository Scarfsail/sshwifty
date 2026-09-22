// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

import * as reader from "../stream/reader.js";
import * as sender from "../stream/sender.js";
import * as integer from "./integer.js";

export class String {
  /**
   * Read String from given reader
   *
   * @param {reader.Reader} rd Source reader
   *
   * @returns {String} read string
   *
   */
  static async read(rd) {
    let l = new integer.Integer(0);
    await l.unmarshall(rd);
    return new String(await reader.readN(rd, l.value()));
  }

  /**
   * constructor
   *
   * @param {Uint8Array} str String data
   */
  constructor(str) {
    this.str = str;
  }

  /**
   * Return the string
   *
   * @returns {Uint8Array} String data
   *
   */
  data() {
    return this.str;
  }

  /**
   * Return serialized String as array
   *
   * @returns {Uint8Array} serialized String
   *
   */
  buffer() {
    let lBytes = new integer.Integer(this.str.length).marshal(),
      buf = new Uint8Array(lBytes.length + this.str.length);
    buf.set(lBytes, 0);
    buf.set(this.str, lBytes.length);
    return buf;
  }
}

/**
 * Truncates a string to the maximum length
 *
 * @param {string} str Source string
 * @param {integer} maxLength Max length
 * @param {string} exceed Text appends the string if it was truncated
 *
 * @returns {string} truncated String
 *
 */
export function truncate(str, maxLength, exceed) {
  if (str.length <= maxLength) {
    return str;
  }
  return str.substring(0, maxLength) + exceed;
}

/**
 * Marshal a list of strings
 *
 * @param {Array<String>} items A list of string
 *
 * @returns {Uint8Array} A list of string
 *
 */
export function marshalStrings(items) {
  let total = new integer.Integer(items.length);
  let result = total.marshal();
  for (let i in items) {
    let buf = items[i].buffer();
    let newBuf = new Uint8Array(result.length + buf.length);
    newBuf.set(result);
    newBuf.set(buf, result.length);
    result = newBuf;
  }
  return result;
}

/**
 * Parses a list of strings
 *
 * @param {reader.Reader} rd Source
 *
 * @returns {Array(String)} A list of string
 *
 */
export async function parseStrings(rd) {
  let total = new integer.Integer(0);
  let items = [];
  await total.unmarshall(rd);
  for (let i = 0; i < total.value(); i++) {
    items.push(await String.read(rd));
  }
  return items;
}

/**
 * Converts Uint8Array to string
 *
 * @param {Uint8Array} a Data to decode
 * @param {string} label Decode label
 *
 * @returns {string} Decoded string
 *
 */
export function toString(a, label) {
  return new TextDecoder(label).decode(a);
}

/**
 * Converts Uint8Array to string
 *
 * @param {string} a string to encode
 *
 * @returns {Uint8Array} Encoded string
 *
 */
export function fromString(a) {
  return new TextEncoder().encode(a);
}
