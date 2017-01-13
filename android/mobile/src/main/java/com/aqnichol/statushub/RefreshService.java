package com.aqnichol.statushub;

import com.aqnichol.watchcomm.CommException;
import com.aqnichol.watchcomm.Sender;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.WearableListenerService;

public class RefreshService extends WearableListenerService {
    @Override
    public void onMessageReceived(MessageEvent messageEvent) {
        if (!messageEvent.getPath().equals("/refresh")) {
            return;
        }
        Sender c = new Sender(getApplicationContext());
        try {
            c.connect();
            c.sendMessage(messageEvent.getSourceNodeId(), "/listing", null);
        } catch (CommException e) {
        } finally {
            c.disconnect();
        }
    }
}
