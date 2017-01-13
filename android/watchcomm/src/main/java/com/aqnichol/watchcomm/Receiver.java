package com.aqnichol.watchcomm;

import android.content.Context;
import android.os.Bundle;

import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.MessageApi;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.Wearable;

import java.util.ArrayList;

/**
 * Receiver provides a synchronous API for polling for
 * messages from an external device.
 */
public class Receiver {
    private GoogleApiClient client;
    private ArrayList<MessageEvent> queue;
    private MessageApi.MessageListener listener;

    /**
     * Create a new Receiver.
     *
     * @param ctx The context to use for google APIs.
     */
    public Receiver(Context ctx) {
        listener = new MessageApi.MessageListener() {
            @Override
            public void onMessageReceived(MessageEvent messageEvent) {
                pushMessage(messageEvent);
            }
        };
        client = new GoogleApiClient.Builder(ctx)
                .addApi(Wearable.API)
                .addConnectionCallbacks(new GoogleApiClient.ConnectionCallbacks() {
                    @Override
                    public void onConnected(Bundle bundle) {
                        Wearable.MessageApi.addListener(client, listener);
                    }

                    @Override
                    public void onConnectionSuspended(int i) {
                        Wearable.MessageApi.removeListener(client, listener);
                    }
                })
                .build();
    }

    /**
     * Connect the receiver asynchronously.
     */
    public void connect() {
        client.connect();
    }

    /**
     * Stop the Receiver until the next connect().
     */
    public void disconnect() {
        if (client.isConnected()) {
            Wearable.MessageApi.removeListener(client, listener);
            client.disconnect();
        }
    }

    /**
     * Remove all pending messages from the incoming message
     * queue, so that poll() will only see messages which
     * were received after clearQueue() was called.
     */
    public synchronized void clearQueue() {
        queue.clear();
    }

    /**
     * Poll receives the next message.
     *
     * @param timeoutMs The timeout in milliseconds.
     * @return The received message or null if the timeout elapsed.
     * @throws CommException If not connected to Play Services.
     */
    public MessageEvent poll(long timeoutMs) throws CommException {
        long startTime = System.currentTimeMillis();
        while (System.currentTimeMillis() > startTime + timeoutMs) {
            MessageEvent e = popMessage();
            if (e != null) {
                return e;
            }
            long remaining = timeoutMs + startTime - System.currentTimeMillis();
            if (remaining < 0) {
                break;
            }
            try {
                wait(remaining);
            } catch (InterruptedException e1) {
            }
        }
        return null;
    }

    private synchronized void pushMessage(MessageEvent m) {
        queue.add(m);
        notifyAll();
    }

    private synchronized MessageEvent popMessage() throws CommException {
        if (!client.isConnected()) {
            throw new CommException("not connected to Play Services");
        }
        if (queue.isEmpty()) {
            return null;
        }
        MessageEvent res = queue.get(0);
        queue.remove(0);
        return res;
    }
}
