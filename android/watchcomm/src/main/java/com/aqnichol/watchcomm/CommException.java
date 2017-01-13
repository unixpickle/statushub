package com.aqnichol.watchcomm;

/**
 * CommException is an error that occurred while trying to
 * communicate between two devices.
 */
public class CommException extends Exception {
    CommException(String msg) {
        super(msg);
    }
}
